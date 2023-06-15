package merchant

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/service/merchantchannelrateservice"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	"context"
	"errors"
	"github.com/gioco-play/easy-i18n/i18n"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantQueryChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantQueryChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantQueryChannelLogic {
	return MerchantQueryChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantQueryChannelLogic) MerchantQueryChannel(req *types.MerchantQueryChannelRequestX) (resp *types.MerchantQueryChannelResponse, err error) {
	var merchant *types.Merchant

	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", req.MerchantId).
		Where("status = ?", "1").
		Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	// 檢查白名單
	if isWhite := merchantsService.IPChecker(req.MyIp, merchant.ApiIP); !isWhite {
		logx.WithContext(l.ctx).Errorf("此IP非法登錄，請設定白名單 来源IP:%s, 白名单:%s", req.MyIp, merchant.ApiIP)
		return nil, errorz.New(response.IP_DENIED, "IP: "+req.MyIp)
	}

	var device string
	if len(req.Device) > 0 {
		device = req.Device
	} else {
		device = "All"
	}

	// 取得对应商户可用渠道
	merchantChannelRates, err := merchantchannelrateservice.GetAllMerchantChannelRateByPayType(l.svcCtx.MyDB, req.MerchantId, req.Currency, req.PayType, device)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("查询商户可用渠道错误, 商户:%s, 错误:%s", req.MerchantId, err.Error())
		return nil, err
	}

	if len(merchantChannelRates) == 0 {
		logx.WithContext(l.ctx).Errorf("查無商戶可用渠道, 商户:%s, 幣別:%s, 支付類型: %s", req.MerchantId, req.Currency, req.PayType)
		return nil, errorz.New(response.DATABASE_FAILURE)
	}

	var merchantQueryChannelPayMethods []types.MerchantQueryChannelPayMethod
	for _, rate := range merchantChannelRates {
		merchantQueryChannelPayMethod := types.MerchantQueryChannelPayMethod{
			CurrencyCoding:     rate.CurrencyCode,
			PayType:            rate.PayTypeCode,
			PayTypeName:        rate.PayTypeName,
			PayTypeImageUrl:    rate.PayTypeImgUrl,
			PayTypeNo:          rate.DesignationNo,
			ChannelCoding:      rate.ChannelCode,
			ChannelName:        rate.ChannelName,
			SingleLimitMaxmum:  rate.SingleMaxCharge,
			SingleLimitMinimum: rate.SingleMinCharge,
			Fixed:              rate.FixedAmount,
			Device:             rate.Device,
		}
		merchantQueryChannelPayMethods = append(merchantQueryChannelPayMethods, merchantQueryChannelPayMethod)
	}

	i18n.SetLang(language.English)
	resp = &types.MerchantQueryChannelResponse{
		RespCode: response.API_SUCCESS,
		RespMsg:  i18n.Sprintf(response.API_SUCCESS),
		Status:   0,
		Data:     merchantQueryChannelPayMethods,
	}
	return resp, nil
}
