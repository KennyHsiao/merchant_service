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
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func PayCallBackHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PayCallBackRequest

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

		l := payorder.NewPayCallBackLogic(r.Context(), ctx)
		// 驗證密鑰
		authenticationPaykey := r.Header.Get("authenticationPaykey")
		if isOK, err := utils.MicroServiceVerification(authenticationPaykey, ctx.Config.ApiKey.PayKey, ctx.Config.ApiKey.PublicKey); err != nil || !isOK {
			err = errorz.New(response.INTERNAL_SIGN_ERROR)
			response.Json(w, r, err.Error(), nil, err)
			return
		}
		resp, err := l.PayCallBack(req)
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
			logx.Errorf("產生交易日志錯誤:%s", errMars.Error())

			if err = transactionLogService.CreateTransactionLog(ctx.MyDB, &types.TransactionLogData{
				//MerchantNo: orderX.MerchantCode,
				//MerchantOrderNo: orderX.MerchantOrderNo,
				OrderNo:   req.PayOrderNo,
				LogType:   constants.ERROR_REPLIED_TO_MERCHANT,
				LogSource: constants.API_ZF,
				Content:   string(contentByte),
				TraceId:   trace.SpanContextFromContext(r.Context()).TraceID().String(),
			}); err != nil {
				logx.WithContext(r.Context()).Errorf("写入交易日志错误:%s", err)
			}
			response.Json(w, r, err.Error(), nil, err)
		} else {
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
