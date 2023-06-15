package proxypayorder

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/logic/proxypayorder"
	transactionLogService "com.copo/bo_service/merchant/internal/service/transactionLog"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func ProxyPayCallBackHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ProxyPayOrderCallBackRequest

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		if err := httpx.ParseJsonBody(r, &req); err != nil {
			response.Json(w, r, response.FAIL, nil, err)
			return
		}

		if err := utils.MyValidator.Struct(req); err != nil {
			response.Json(w, r, response.INVALID_PARAMETER, nil, err)
			return
		}

		if requestBytes, err := json.Marshal(req); err == nil {
			span.SetAttributes(attribute.KeyValue{
				Key:   "request",
				Value: attribute.StringValue(string(requestBytes)),
			})
		}

		l := proxypayorder.NewProxyPayCallBackLogic(r.Context(), ctx)
		resp, err := l.ProxyPayCallBack(&req)
		if err != nil {
			// 写入交易日志
			contentStrut := struct {
				Error    string
				ErrorMsg string
			}{
				Error:    "代付渠道回調錯誤",
				ErrorMsg: err.Error(),
			}
			contentByte, errMars := json.Marshal(contentStrut)
			if errMars != nil {
				logx.Errorf("產生交易日志錯誤:%s", errMars.Error())
			}

			if errLog := transactionLogService.CreateTransactionLog(ctx.MyDB, &types.TransactionLogData{
				//MerchantNo:      "",
				//MerchantOrderNo: req.ProxyPayOrderNo,
				OrderNo:   req.ProxyPayOrderNo,
				LogType:   constants.ERROR_MSG,
				LogSource: constants.API_DF,
				Content:   string(contentByte),
				TraceId:   trace.SpanContextFromContext(r.Context()).TraceID().String(),
			}); errLog != nil {
				logx.WithContext(r.Context()).Errorf("写入交易日志错误:%s", errLog)
			}
			response.Json(w, r, err.Error(), resp, err)
		} else {
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
