package merchant

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	"context"
	"errors"
	"github.com/copo888/transaction_service/common/constants"
	"github.com/gioco-play/easy-i18n/i18n"
	"gorm.io/gorm"
	"time"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantQueryOrderStatementLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantQueryOrderStatementLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantQueryOrderStatementLogic {
	return MerchantQueryOrderStatementLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantQueryOrderStatementLogic) MerchantQueryOrderStatement(req *types.MerchantQueryOrderStatementRequestX) (resp *types.MerchantQueryOrderStatementResponse, err error) {
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
	if isWhite := merchantsService.IPChecker(req.Ip, merchant.ApiIP); !isWhite {
		logx.WithContext(l.ctx).Errorf("此IP非法登錄，請設定白名單 来源IP:%s, 白名单:%s", req.Ip, merchant.ApiIP)
		return nil, errorz.New(response.IP_DENIED, "IP: "+req.Ip)
	}

	// 验签检查
	if isSameSign := utils.VerifySign(req.Sign, req.MerchantQueryOrderStatementRequest, merchant.ScrectKey, l.ctx); !isSameSign {
		logx.WithContext(l.ctx).Errorf("签名出错")
		return nil, errorz.New(response.INVALID_SIGN)
	}

	st, err1 := time.Parse("2006-01-02 15:04:05", req.StartAt)
	if err1 != nil {
		return nil, err1
	}

	ed, err1 := time.Parse("2006-01-02 15:04:05", req.EndAt)
	if err1 != nil {
		return nil, err1
	}

	t := ed.Sub(st).Hours() / 24

	if t > 31 {
		logx.WithContext(l.ctx).Errorf("查询商户对帐，时间区间大于31天")
		return nil, errorz.New(response.INVALID_DATE_RANGE)
	}

	db := l.svcCtx.MyDB
	db = db.Where("tx.merchant_code = ?", req.MerchantId).
		Where("tx.currency_code = ?", req.Currency).
		Where("tx.status in ?", []string{constants.SUCCESS, constants.FROZEN}).
		Where("tx.is_test != ?", constants.IS_TEST_YES).
		Where("tx.reason_type != ?", constants.ORDER_REASON_TYPE_RECOVER)

	if len(req.StartAt) > 0 {
		timeString, err1 := time.Parse("2006-01-02 15:04:05", req.StartAt)
		if err1 != nil {
			return nil, err1
		}
		str := timeString.Add(time.Hour * -8).Format("2006-01-02 15:04:05")
		db = db.Where("tx.`trans_at` >= ?", str)
	}

	if len(req.EndAt) > 0 {
		timeString, err1 := time.Parse("2006-01-02 15:04:05", req.EndAt)
		if err1 != nil {
			return nil, err1
		}
		str := timeString.Add(time.Hour * -8).Format("2006-01-02 15:04:05")
		db = db.Where("tx.`trans_at` < ?", str)
	}

	selectX := "COUNT(if (tx.type = 'ZF', true, null)) as pay_order_num," +
		"IFNULL(SUM(if (tx.type = 'ZF', tx.actual_amount, 0)),0) as pay_order_amount," +
		"IFNULL(SUM(if (tx.type = 'ZF', tx.transfer_handling_fee, 0)),0) as pay_order_handling_fee," +
		"COUNT(if (tx.type = 'DF',true, null)) as proxy_pay_order_num," +
		"IFNULL(SUM(if (tx.type = 'DF', tx.order_amount, 0)),0) as proxy_pay_order_amount," +
		"IFNULL(SUM(if (tx.type = 'DF', tx.transfer_handling_fee, 0)),0) as proxy_pay_order_handling_fee"

	if err := l.svcCtx.MyDB.Table("tx_orders tx").Select(selectX).Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE)
	}

	resp.Sign = utils.SortAndSign2(*resp, merchant.ScrectKey)
	resp.RespCode = response.API_SUCCESS
	resp.RespMsg = i18n.Sprintf(response.API_SUCCESS)
	return
}
