package payorder

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/logic/payorder"
	transactionLogService "com.copo/bo_service/merchant/internal/service/transactionLog"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"encoding/json"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/thinkeridea/go-extend/exnet"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
)

func PayOrderHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PayOrderRequestX
		r.WithContext(r.Context())
		span := trace.SpanFromContext(r.Context())
		defer span.End()

		bodyBytes, err := io.ReadAll(r.Body)

		if err != nil {
			response.Json(w, r, response.FAIL, nil, err)
			return
		}

		logx.WithContext(r.Context()).Infof("PayOrder enter: %s", string(bodyBytes))

		if err = json.Unmarshal(bodyBytes, &req); err != nil {
			errx := errorz.New(response.API_INVALID_PARAMETER, err.Error())
			response.ApiErrorJson(w, r, errx.Error(), errx)
			return
		}

		// 写入交易日志
		if errLog := transactionLogService.CreateTransactionLog(ctx.MyDB, &types.TransactionLogData{
			MerchantNo:      req.MerchantId,
			MerchantOrderNo: req.OrderNo,
			//OrderNo:         orderNo,
			LogType:   constants.MERCHANT_REQUEST,
			LogSource: constants.API_ZF,
			Content:   req,
			TraceId:   trace.SpanContextFromContext(r.Context()).TraceID().String(),
		}); errLog != nil {
			logx.WithContext(r.Context()).Errorf("写入交易日志错误: %s", errLog.Error())
		}

		if err := utils.MyValidator.Struct(req); err != nil {
			response.ApiErrorJson(w, r, response.API_INVALID_PARAMETER, err)
			return
		}

		myIP := exnet.ClientIP(r)
		req.MyIp = myIP

		if requestBytes, err := json.Marshal(req); err == nil {
			span.SetAttributes(attribute.KeyValue{
				Key:   "request",
				Value: attribute.StringValue(string(requestBytes)),
			})
		}

		l := payorder.NewPayOrderLogic(r.Context(), ctx)
		resp, err := l.PayOrder(req)
		payOrderNo := ""
		if resp != nil {
			payOrderNo = resp.PayOrderNo
		}
		if err != nil {
			var msg string
			if v, ok := err.(*errorz.Err); ok && v.GetMessage() != "" {
				msg = v.GetMessage()
			} else {
				msg = i18n.Sprintf("%s", err.Error())
			}
			// 写入交易日志
			contentStrut := struct {
				ErrorCode string
				ErrorMsg  string
				//Resp      interface{}
			}{
				ErrorCode: fmt.Sprintf("%s", err.Error()),
				ErrorMsg:  msg,
				//Resp:      resp,
			}

			if errLog := transactionLogService.CreateTransactionLog(ctx.MyDB, &types.TransactionLogData{
				MerchantNo:      req.MerchantId,
				MerchantOrderNo: req.OrderNo,
				OrderNo:         payOrderNo,
				LogType:         constants.ERROR_REPLIED_TO_MERCHANT,
				LogSource:       constants.API_ZF,
				Content:         contentStrut,
				TraceId:         trace.SpanContextFromContext(r.Context()).TraceID().String(),
			}); errLog != nil {
				logx.WithContext(r.Context()).Errorf("写入交易日志错误:%s", errLog)
			}
			response.ApiErrorJson(w, r, err.Error(), err)
		} else {
			response.ApiJson(w, r, resp)
		}
	}
}
