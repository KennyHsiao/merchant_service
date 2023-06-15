package payorder

import (
	"com.copo/bo_service/common/constants/redisKey"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/model"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	ordersService "com.copo/bo_service/merchant/internal/service/orders"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/gioco-play/gozzle"
	"github.com/jinzhu/copier"
	"github.com/neccoys/go-zero-extension/redislock"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/text/language"
	"strconv"
	"strings"
	"time"

	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type PayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayOrderLogic {
	return PayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayOrderLogic) PayOrder(req types.PayOrderRequestX) (resp *types.PayOrderResponse, err error) {

	redisKey := fmt.Sprintf("%s-%s", req.MerchantId, req.OrderNo)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "pay-order:")
	redisLock.SetExpire(5)

	if isOK, redisErr := redisLock.Acquire(); redisErr != nil {
		logx.WithContext(l.ctx).Errorf("redisLock error: %s", redisErr.Error())
		return resp, errorz.New(response.TRANSACTION_PROCESSING)
	} else if isOK {
		if resp, err = l.DoPayOrder(req); err != nil {
			return
		}
		defer redisLock.Release()
	} else {
		return resp, errorz.New(response.TRANSACTION_PROCESSING)
	}

	return
}

func (l *PayOrderLogic) DoPayOrder(req types.PayOrderRequestX) (resp *types.PayOrderResponse, err error) {

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

	req.PayOrderRequest.OrderAmount = req.OrderAmount.String()
	req.PayOrderRequest.AccessType = req.AccessType.String()
	req.PayOrderRequest.PayTypeNo = req.PayTypeNo.String()

	// 检查是否商户有启用此币别
	if errCurrency := l.svcCtx.MyDB.Table("mc_merchant_currencies").
		Where("merchant_code = ? AND currency_code = ? AND status = ?", req.MerchantId, req.Currency, "1").
		Take(map[string]interface{}{}).
		Error; errCurrency != nil {
		return nil, errorz.New(response.MERCHANT_CURRENCY_NOT_SET)
	}

	if isSameSign := utils.VerifySign(req.Sign, req.PayOrderRequest, merchant.ScrectKey, l.ctx); !isSameSign {
		logx.WithContext(l.ctx).Errorf("签名出错")
		return nil, errorz.New(response.INVALID_SIGN)
	}

	// 资料验证
	if err = ordersService.VerifyPayOrder(l.svcCtx.MyDB, req, merchant); err != nil {
		logx.WithContext(l.ctx).Errorf("VerifyPayOrder error:%s", err.Error())
		return
	}

	// 產生訂單號
	orderNo := model.GenerateOrderNo("ZF")
	resp = &types.PayOrderResponse{
		PayOrderNo: orderNo,
	}

	// 確認是否返回實名制UI畫面
	if req.JumpType == "UI" {
		resp, err = l.RequireUserIdPage(req, orderNo)
		return
	}

	// 若商户没给PageUrl 则给我们的预设的跳转完成页
	if req.PageUrl == "" {
		req.PageUrl = fmt.Sprintf("%s/#/messagePage", l.svcCtx.Config.FrontEndDomain)
	}

	// Call Channel
	payReplyVO, correspondMerChnRate, chaErr := ordersService.CallChannelForPay(l.svcCtx.MyDB, req, merchant, orderNo, l.ctx, l.svcCtx)
	if chaErr != nil {
		logx.WithContext(l.ctx).Errorf("CallChannelForPay error:%s", chaErr.Error())
		return resp, chaErr
	}

	// Call GRPC transaction_service
	var rpcPayOrder transaction.PayOrder
	var rpcRate transaction.CorrespondMerChnRate
	copier.Copy(&rpcPayOrder, &req)
	copier.Copy(&rpcRate, correspondMerChnRate)
	rpcPayOrder.OrderAmount = req.OrderAmount.String()

	// CALL transactionc PayOrderTranaction
	rpcResp, err2 := l.svcCtx.TransactionRpc.PayOrderTranaction(l.ctx, &transaction.PayOrderRequest{
		PayOrder:       &rpcPayOrder,
		Rate:           &rpcRate,
		OrderNo:        orderNo,
		ChannelOrderNo: payReplyVO.ChannelOrderNo,
	})
	if err2 != nil {
		logx.WithContext(l.ctx).Errorf("PayOrderTranaction rpcResp error:%s", err2.Error())
		return resp, err2
	} else if rpcResp == nil {
		logx.WithContext(l.ctx).Errorf("Code:%s, Message:%s", rpcResp.Code, rpcResp.Message)
		return resp, errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "PayOrderTranaction rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		logx.WithContext(l.ctx).Errorf("Code:%s, Message:%s", rpcResp.Code, rpcResp.Message)
		return resp, errorz.New(rpcResp.Code, rpcResp.Message)
	}

	// 判斷返回格式 1.html, 2.json  3.url
	resp, err = ordersService.GetPayOrderResponse(req, *payReplyVO, orderNo, l.ctx, l.svcCtx)
	if err != nil {
		return
	}
	i18n.SetLang(language.English)
	resp.RespCode = response.API_SUCCESS
	resp.RespMsg = i18n.Sprintf(response.API_SUCCESS)
	resp.Status = 0
	resp.Sign = utils.SortAndSign2(*resp, merchant.ScrectKey)

	l.CheckForConsecutiveUnpaidOrdersToNotify(req, merchant)

	return
}

func (l *PayOrderLogic) RequireUserIdPage(req types.PayOrderRequestX, orderNo string) (*types.PayOrderResponse, error) {
	// 资料转JSON
	dataJson, err := json.Marshal(req)
	if err != nil {
		return nil, errorz.New(response.API_PARAMETER_TYPE_ERROE)
	}
	// 存 Redis
	if err = l.svcCtx.RedisClient.Set(l.ctx, redisKey.CACHE_ORDER_DATA+orderNo, dataJson, 30*time.Minute).Err(); err != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION)
	}

	url := fmt.Sprintf("%s/#/checkoutPlayer?id=%s&lang=%s", l.svcCtx.Config.FrontEndDomain, orderNo, req.Language)

	return &types.PayOrderResponse{
		Status:     0,
		PayOrderNo: orderNo,
		BankCode:   req.BankCode,
		Type:       "url",
		Info:       url,
	}, nil
}

func (l *PayOrderLogic) CheckForConsecutiveUnpaidOrdersToNotify(req types.PayOrderRequestX, merchant *types.Merchant) {
	go func() {
		l.DoCheckForConsecutiveUnpaidOrdersToNotify(req, merchant)
	}()
}

func (l *PayOrderLogic) DoCheckForConsecutiveUnpaidOrdersToNotify(req types.PayOrderRequestX, merchant *types.Merchant) {
	var orderQuantity int
	var timeMinute int64

	var unpaidNotifyIntervalParam types.SystemParams
	var unpaidNotifyNumParam types.SystemParams
	var unpaidNotifyPayTypeParam types.SystemParams
	var orders []types.OrderX
	var err error

	// 取得要查刷空單的支付類型
	if err = l.svcCtx.MyDB.Table("bs_system_params").
		Select("value").
		Where("name = 'unpaidNotifyPayType'").
		Take(&unpaidNotifyPayTypeParam).Error; err != nil {
		logx.WithContext(l.ctx).Errorf("空單通知: unpaidNotifyPayType 參數取得失敗:%s", err.Error())
		return
	}
	if strings.Index(unpaidNotifyPayTypeParam.Value, req.PayType) < 0 {
		// 支付類型不符 不用檢查空單
		return
	}

	if merchant.UnpaidNotifyNum > 0 {
		orderQuantity = int(merchant.UnpaidNotifyNum)
	} else {
		// 商戶沒有客製設定 取得系統參數
		// 空單檢查訂單數量
		if err := l.svcCtx.MyDB.Table("bs_system_params").
			Where("name = 'unpaidNotifyNum'").
			Take(&unpaidNotifyNumParam).Error; err != nil {
			logx.WithContext(l.ctx).Errorf("空單通知: unpaidNotifyNum 參數取得失敗:%s", err.Error())
			return
		}
		if orderQuantity, err = strconv.Atoi(unpaidNotifyNumParam.Value); err != nil {
			logx.WithContext(l.ctx).Errorf("參數轉換失敗:%s", err.Error())
			return
		}
	}

	if merchant.UnpaidNotifyInterval > 0 {
		timeMinute = merchant.UnpaidNotifyInterval
	} else {
		// 商戶沒有客製設定 取得系統參數
		// 取得空單時間週期(分)參數
		if err := l.svcCtx.MyDB.Table("bs_system_params").
			Select("value").
			Where("name = 'unpaidNotifyInterval'").
			Take(&unpaidNotifyIntervalParam).Error; err != nil {
			logx.WithContext(l.ctx).Errorf("空單通知: unpaidNotifyInterval 參數取得失敗:%s", err.Error())
			return
		}

		if timeMinute, err = strconv.ParseInt(unpaidNotifyIntervalParam.Value, 10, 64); err != nil {
			logx.WithContext(l.ctx).Errorf("參數轉換失敗:%s", err.Error())
			return
		}
	}

	if orderQuantity <= 0 {
		logx.WithContext(l.ctx).Errorf("參數異常 orderQuantity:%s", orderQuantity)
		return
	}

	if timeMinute <= 0 {
		logx.WithContext(l.ctx).Errorf("參數異常 timeMinute:%s", timeMinute)
		return
	}

	startTime := time.Now().Add(-time.Minute * time.Duration(timeMinute))

	if err := l.svcCtx.MyDB.Table("tx_orders").
		Where("type = ?", "ZF").
		Where("merchant_code = ?", req.MerchantId).
		Where("pay_type_code = ?", req.PayType).
		Where("currency_code = ?", req.Currency).
		Where("created_at >= ?", startTime).
		Order("created_at desc").
		Limit(orderQuantity).
		Find(&orders).Error; err != nil {
		logx.WithContext(l.ctx).Errorf("空單通知: 訂單取得失敗:%s", err.Error())
		return
	}

	if len(orders) < orderQuantity {
		return
	}

	for _, order := range orders {
		// 週期內有成功單不需要通知
		if order.Status == "20" {
			return
		}
	}
	h, _ := time.ParseDuration("1h")
	notifyBO := types.TelegramNotifyRequest{
		Message: fmt.Sprintf("商户号： %s\n支付类型： %s\n%s\n到%s\n發生刷空單狀況",
			req.MerchantId,
			req.PayType,
			orders[len(orders)-1].CreatedAt.Add(8*h).Format("2006-01-02 15:04:05"),
			orders[0].CreatedAt.Add(8*h).Format("2006-01-02 15:04:05")),
	}
	url := fmt.Sprintf("%s:20003/telegram/notify", l.svcCtx.Config.Server)
	span := trace.SpanFromContext(l.ctx)
	if _, err := gozzle.Post(url).Timeout(25).Trace(span).JSON(notifyBO); err != nil {
		logx.WithContext(l.ctx).Errorf("空單通知失敗:%s", err.Error())
	}

}
