package payorder

import (
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/model"
	transactionLogService "com.copo/bo_service/merchant/internal/service/transactionLog"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"errors"
	"fmt"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/gioco-play/gozzle"
	"github.com/neccoys/go-zero-extension/redislock"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayCallBackLogic {
	return PayCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayCallBackLogic) PayCallBack(req types.PayCallBackRequest) (resp *types.PayCallBackResponse, err error) {

	// 只能回調成功/失敗
	if req.OrderStatus != "20" && req.OrderStatus != "30" {
		logx.WithContext(l.ctx).Infof("支付订单回调[非成功/失败状态]: 单号:%s, 状态:%s", req.PayOrderNo, req.OrderStatus)
		return
	}

	//检查单号是否存在
	orderX := &types.OrderX{}
	if req.PayOrderNo == "" {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST)
	} else if orderX, err = model.QueryOrderByOrderNo(l.svcCtx.MyDB, req.PayOrderNo, ""); err != nil && orderX == nil {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST, "Copo OrderNo: "+req.PayOrderNo)
	} else if orderX.Status == "20" || orderX.Status == "30" {
		//已经回调成功过
		return
	}

	// 写入交易日志
	var errLog error
	if errLog = transactionLogService.CreateTransactionLog(l.svcCtx.MyDB, &types.TransactionLogData{
		MerchantNo:      orderX.MerchantCode,
		MerchantOrderNo: orderX.MerchantOrderNo,
		OrderNo:         orderX.OrderNo,
		LogType:         constants.CALLBACK_FROM_CHANNEL,
		LogSource:       constants.API_ZF,
		Content:         req,
		TraceId:         trace.SpanContextFromContext(l.ctx).TraceID().String(),
	}); errLog != nil {
		logx.WithContext(l.ctx).Errorf("写入交易日志错误:%s", errLog)
	}

	redisKey := fmt.Sprintf("%s-%s", orderX.MerchantCode, orderX.OrderNo)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "pay-call-back:")
	redisLock.SetExpire(5)
	if isOK, _ := redisLock.Acquire(); isOK {
		defer redisLock.Release()
		return l.DoPayCallBack(req, orderX)
	} else {
		logx.WithContext(l.ctx).Infof(i18n.Sprintf(response.TRANSACTION_PROCESSING))
		return nil, errorz.New(response.TRANSACTION_PROCESSING)
	}
	return
}

func (l *PayCallBackLogic) DoPayCallBack(req types.PayCallBackRequest, orderX *types.OrderX) (resp *types.PayCallBackResponse, err error) {

	// CALL transactionc PayOrderTranaction
	callBackResp, err2 := l.svcCtx.TransactionRpc.PayCallBackTranaction(l.ctx, &transaction.PayCallBackRequest{
		CallbackTime:   req.CallbackTime,
		ChannelOrderNo: req.ChannelOrderNo,
		OrderAmount:    req.OrderAmount,
		OrderStatus:    req.OrderStatus,
		PayOrderNo:     req.PayOrderNo,
	})

	if err2 != nil {
		logx.WithContext(l.ctx).Errorf("PayCallBackTranaction失敗: %s", err2.Error())
		return nil, err2
	} else if callBackResp == nil {
		logx.WithContext(l.ctx).Errorf("PayCallBackTranaction失敗: %s", "PayCallBackTranaction callBackResp is nil")
		return nil, errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "PayCallBackTranaction callBackResp is nil")
	} else if callBackResp.Code != response.API_SUCCESS {
		logx.WithContext(l.ctx).Errorf("PayCallBackTranaction失敗: %s", callBackResp.Message)
		return nil, errorz.New(callBackResp.Code, callBackResp.Message)
	}

	logx.WithContext(l.ctx).Infof("PayCallBackTranaction return: %#v", callBackResp)

	// 只有成功/失敗單 且 有回掉網址 才回調
	if (req.OrderStatus == "20" || req.OrderStatus == "30") && len(callBackResp.NotifyUrl) > 0 {
		// 异步回调商户
		go func() {
			l.callNoticeURL(callBackResp)
		}()
	}

	return
}

func (l *PayCallBackLogic) callNoticeURL(callBackResp *transaction.PayCallBackResponse) (respString string, err error) {
	// 会调状态改为必成功
	if err4 := l.svcCtx.MyDB.Table("tx_orders").
		Where("order_no = ?", callBackResp.OrderNo).
		Updates(map[string]interface{}{"is_merchant_callback": "1"}).Error; err4 != nil {
		logx.WithContext(l.ctx).Error("PayCallback To Merchant 改回調狀態失敗")
	}
	var merchant *types.Merchant
	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", callBackResp.MerchantCode).
		Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logx.WithContext(l.ctx).Errorf("callNoticeURL失敗: %s, %s", respString, err.Error())
			return "", errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			logx.WithContext(l.ctx).Errorf("callNoticeURL失敗: %s, %s", respString, err.Error())
			return "", errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	payCallBackVO := vo.PayCallBackVO{
		AccessType:   "1",
		Language:     "zh-CN",
		MerchantId:   callBackResp.MerchantCode,
		OrderNo:      callBackResp.MerchantOrderNo,
		OrderTime:    callBackResp.OrderTime,
		PayOrderTime: callBackResp.PayOrderTime,
		Fee:          fmt.Sprintf("%.2f", callBackResp.TransferHandlingFee),
		PayOrderId:   callBackResp.OrderNo,
	}

	// 若有實際金額則回覆實際
	payCallBackVO.OrderAmount = fmt.Sprintf("%.2f", callBackResp.ActualAmount)

	// API 支付状态 0：处理中，1：成功，2：失败，3：成功(人工确认)
	payCallBackVO.OrderStatus = "0"
	if callBackResp.Status == constants.SUCCESS {
		payCallBackVO.OrderStatus = "1"
	} else if callBackResp.Status == constants.FAIL {
		payCallBackVO.OrderStatus = "2"
	}

	payCallBackVO.Sign = utils.SortAndSign2(payCallBackVO, merchant.ScrectKey)

	var minDelaySeconds int64 = 10
	for i := 0; i < 5; i++ {
		startTime := time.Now().Unix()
		logx.WithContext(l.ctx).Infof("PayCallback To Merchant: 第%d次回調 訂單:%s, NotifyUrl:%s, request: %+v", i+1, payCallBackVO.PayOrderId, callBackResp.NotifyUrl, payCallBackVO)
		if isOk := l.DoCallNoticeURL(callBackResp.NotifyUrl, payCallBackVO); isOk {
			logx.WithContext(l.ctx).Infof("PayCallback To Merchant: 訂單:%s 回調成功", payCallBackVO.PayOrderId)

			break
		}
		endTime := time.Now().Unix()
		secondsDiff := endTime - startTime
		if secondsDiff < minDelaySeconds {
			sleepTime := time.Duration(minDelaySeconds-secondsDiff) * time.Second
			time.Sleep(sleepTime)
		}
	}

	return
}

func (l *PayCallBackLogic) DoCallNoticeURL(notifyUrl string, payCallBackVO vo.PayCallBackVO) (isOk bool) {

	// 写入交易日志
	if err := transactionLogService.CreateTransactionLog(l.svcCtx.MyDB, &types.TransactionLogData{
		MerchantNo:      payCallBackVO.MerchantId,
		MerchantOrderNo: payCallBackVO.OrderNo,
		OrderNo:         payCallBackVO.PayOrderId,
		LogType:         constants.CALLBACK_TO_MERCHANT,
		LogSource:       constants.API_ZF,
		Content:         payCallBackVO,
		TraceId:         trace.SpanContextFromContext(l.ctx).TraceID().String(),
	}); err != nil {
		logx.WithContext(l.ctx).Errorf("写入交易日志错误:%s", err)
	}

	// 通知商戶
	span := trace.SpanFromContext(l.ctx)
	res, errx := gozzle.Post(notifyUrl).Timeout(20).Trace(span).JSON(payCallBackVO)
	if errx != nil {
		logx.WithContext(l.ctx).Errorf("PayCallback To Merchant gozzle Error: %s, Error:%s", payCallBackVO.PayOrderId, errx.Error())
		return false
	}
	resString := string(res.Body()[:])
	if res.Status() < 200 || res.Status() >= 300 {
		logx.WithContext(l.ctx).Errorf("PayCallback To Merchant 状态码错误: %s, HttpStatus:%d, Response:%s",
			payCallBackVO.PayOrderId, res.Status(), resString)
		return false
	}

	// 写入交易日志
	var errLog error
	if errLog = transactionLogService.CreateTransactionLog(l.svcCtx.MyDB, &types.TransactionLogData{
		MerchantNo:      payCallBackVO.MerchantId,
		MerchantOrderNo: payCallBackVO.OrderNo,
		OrderNo:         payCallBackVO.PayOrderId,
		LogType:         constants.RESPONSE_FROM_MERCHANT,
		LogSource:       constants.API_ZF,
		Content:         resString,
		TraceId:         trace.SpanContextFromContext(l.ctx).TraceID().String(),
	}); errLog != nil {
		logx.WithContext(l.ctx).Errorf("写入交易日志错误:%s", errLog)
	}

	if resString == "success" {
		logx.WithContext(l.ctx).Errorf("PayCallback To Merchant 回調成功: %s, Response:%s", payCallBackVO.PayOrderId, resString)
		return true
	} else {
		logx.WithContext(l.ctx).Errorf("PayCallback To Merchant 商户回复错误: %s, Response:%s", payCallBackVO.PayOrderId, resString)
		return false
	}
}
