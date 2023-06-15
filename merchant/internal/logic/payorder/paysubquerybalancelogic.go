package payorder

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	"context"
	"errors"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"gorm.io/gorm"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PaySubQueryBalanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPaySubQueryBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) PaySubQueryBalanceLogic {
	return PaySubQueryBalanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PaySubQueryBalanceLogic) PaySubQueryBalance(req types.PayQueryBalanceRequestX) (resp *types.PayQuerySubBalanceResponse, err error) {
	var merchant *types.Merchant
	var merchantPtBalances []types.MerchantPtBalance
	var dfBalance *types.MerchantBalance
	var xfBalance *types.MerchantBalance
	var subBalances []types.SubBalance

	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", req.MerchantId).
		Where("status = ?", "1").
		Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			return nil, errorz.New(response.GENERAL_EXCEPTION, err.Error())
		}
	}

	// 檢查白名單
	if isWhite := merchantsService.IPChecker(req.MyIp, merchant.ApiIP); !isWhite {
		return nil, errorz.New(response.IP_DENIED, "IP: "+req.MyIp)
	}

	req.PayQueryBalanceRequest.AccessType = req.AccessType.String()

	// 檢查驗簽
	if isSameSign := utils.VerifySign(req.Sign, req.PayQueryBalanceRequest, merchant.ScrectKey, l.ctx); !isSameSign {
		return nil, errorz.New(response.INVALID_SIGN)
	}

	if len(req.Currency) < 2 {
		req.Currency = "CNY"
	}

	if err = l.svcCtx.MyDB.Table("mc_merchant_balances").
		Where("merchant_code = ?", req.MerchantId).
		Where("currency_code = ?", req.Currency).
		Where("balance_type =  'DFB' ").
		Take(&dfBalance).Error; err != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}
	if err = l.svcCtx.MyDB.Table("mc_merchant_balances").
		Where("merchant_code = ?", req.MerchantId).
		Where("currency_code = ?", req.Currency).
		Where("balance_type =  'XFB' ").
		Take(&xfBalance).Error; err != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}

	if err = l.svcCtx.MyDB.Table("mc_merchant_pt_balances").
		Where("merchant_code = ?", req.MerchantId).
		Where("currency_code = ?", req.Currency).
		Find(&merchantPtBalances).Error; err != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}

	if len(merchantPtBalances) > 0 {
		for _, balance := range merchantPtBalances {
			var subBalance types.SubBalance
			subBalance.ID = balance.ID
			subBalance.Name = balance.Name
			subBalance.Balance = fmt.Sprintf("%.2f", balance.Balance)
			subBalances = append(subBalances, subBalance)
		}
	}

	resp = &types.PayQuerySubBalanceResponse{
		RespCode:        response.API_SUCCESS,
		ResMsg:          i18n.Sprintf(response.API_SUCCESS),
		FrozenAmount:    fmt.Sprintf("%.2f", utils.FloatAdd(xfBalance.FrozenAmount, dfBalance.FrozenAmount)),
		AvailableAmount: fmt.Sprintf("%.2f", utils.FloatAdd(xfBalance.Balance, dfBalance.Balance)),
		SubBalances:     subBalances,
	}
	resp.Sign = utils.SortAndSign2(*resp, merchant.ScrectKey)
	return
}
