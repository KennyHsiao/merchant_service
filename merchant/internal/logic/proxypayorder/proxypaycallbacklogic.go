package proxypayorder

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/model"
	ordersService "com.copo/bo_service/merchant/internal/service/orders"
	transactionLogService "com.copo/bo_service/merchant/internal/service/transactionLog"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/text/language"
	"strings"
)

type ProxyPayCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyPayCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyPayCallBackLogic {
	return ProxyPayCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyPayCallBackLogic) ProxyPayCallBack(req *types.ProxyPayOrderCallBackRequest) (resp *types.ProxyPayOrderCallBackResponse, err error) {
	logx.WithContext(l.ctx).Infof("渠道回調請求參數: %+v", req)

	//若是撥款單 走專屬流程
	if strings.Index(req.ProxyPayOrderNo, "BK") == 0 {
		return l.ProxyPayCallBackForBK(req)
	}

	//检查单号是否存在
	orderX := &types.OrderX{}
	if req.ProxyPayOrderNo == "" && req.ChannelOrderNo == "" {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST)
	} else if orderX, err = model.QueryOrderByOrderNo(l.svcCtx.MyDB, req.ProxyPayOrderNo, ""); err != nil && orderX == nil {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST, "Copo OrderNo: "+req.ProxyPayOrderNo)
	}

	// 写入交易日志
	var errLog error
	if errLog = transactionLogService.CreateTransactionLog(l.svcCtx.MyDB, &types.TransactionLogData{
		MerchantNo:      orderX.MerchantCode,
		MerchantOrderNo: orderX.MerchantOrderNo,
		OrderNo:         orderX.OrderNo,
		LogType:         constants.CALLBACK_FROM_CHANNEL,
		LogSource:       constants.API_DF,
		Content:         req,
		TraceId:         trace.SpanContextFromContext(l.ctx).TraceID().String(),
	}); errLog != nil {
		logx.WithContext(l.ctx).Errorf("写入交易日志错误:%s", errLog)
	}

	ProxyPayOrderCallBackResp := &types.ProxyPayOrderCallBackResponse{}
	service := ordersService.NewOrdersService(l.ctx, l.svcCtx)
	if errCallBack := service.ChannelCallBackForProxy(req, orderX); errCallBack != nil {
		ProxyPayOrderCallBackResp.RespMsg = errCallBack.Error()
		return ProxyPayOrderCallBackResp, errorz.New(response.PROXY_PAY_CALLBACK_FAIL, errCallBack.Error())
	}

	i18n.SetLang(language.English)
	callBackResp := &types.ProxyPayOrderCallBackResponse{
		RespCode: response.API_SUCCESS,
		RespMsg:  i18n.Sprintf(response.API_SUCCESS),
	}
	return callBackResp, nil
}

func (l *ProxyPayCallBackLogic) getOrderStatus(channelResultStatus string) string {

	var orderStatus string
	switch channelResultStatus {
	case "0":
		orderStatus = constants.TRANSACTION
	case "1":
		orderStatus = constants.TRANSACTION
	case "2":
		orderStatus = constants.SUCCESS
	case "3":
		orderStatus = constants.FAIL
	default:
		orderStatus = constants.TRANSACTION
	}
	return orderStatus
}

func (l *ProxyPayCallBackLogic) ProxyPayCallBackForBK(req *types.ProxyPayOrderCallBackRequest) (*types.ProxyPayOrderCallBackResponse, error) {
	orderX := &types.OrderX{}
	// 確認訂單是否存在
	if err := l.svcCtx.MyDB.Table("tx_orders").
		Where("order_no = ?", req.ProxyPayOrderNo).
		Take(orderX).Error; err != nil {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST, err.Error())
	}

	// 写入交易日志
	var errLog error
	if errLog = transactionLogService.CreateTransactionLog(l.svcCtx.MyDB, &types.TransactionLogData{
		MerchantNo:      orderX.MerchantCode,
		MerchantOrderNo: orderX.MerchantOrderNo,
		OrderNo:         orderX.OrderNo,
		LogType:         constants.CALLBACK_FROM_CHANNEL,
		LogSource:       constants.PLATEFORM_DF,
		Content:         req,
		TraceId:         trace.SpanContextFromContext(l.ctx).TraceID().String(),
	}); errLog != nil {
		logx.WithContext(l.ctx).Errorf("写入交易日志错误:%s", errLog)
	}

	ProxyPayOrderCallBackResp := &types.ProxyPayOrderCallBackResponse{}
	service := ordersService.NewOrdersService(l.ctx, l.svcCtx)
	if errCallBack := service.ChannelCallBackAlloc(req, orderX); errCallBack != nil {
		ProxyPayOrderCallBackResp.RespMsg = errCallBack.Error()
		return ProxyPayOrderCallBackResp, errorz.New(response.PROXY_PAY_CALLBACK_FAIL, errCallBack.Error())
	}

	i18n.SetLang(language.English)
	callBackResp := &types.ProxyPayOrderCallBackResponse{
		RespCode: response.API_SUCCESS,
		RespMsg:  i18n.Sprintf(response.API_SUCCESS),
	}
	return callBackResp, nil
}
