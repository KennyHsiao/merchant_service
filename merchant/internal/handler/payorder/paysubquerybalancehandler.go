package payorder

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/logic/payorder"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"encoding/json"
	"github.com/thinkeridea/go-extend/exnet"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
)

func PaySubQueryBalanceHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PayQueryBalanceRequestX

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		bodyBytes, err := io.ReadAll(r.Body)

		if err != nil {
			response.Json(w, r, response.FAIL, nil, err)
			return
		}

		logx.WithContext(r.Context()).Infof("PayQueryBalance enter: %s", string(bodyBytes))

		if err = json.Unmarshal(bodyBytes, &req); err != nil {
			errx := errorz.New(response.API_INVALID_PARAMETER, err.Error())
			response.ApiErrorJson(w, r, errx.Error(), errx)
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

		myIP := exnet.ClientIP(r)
		req.MyIp = myIP

		l := payorder.NewPaySubQueryBalanceLogic(r.Context(), ctx)
		resp, err := l.PaySubQueryBalance(req)
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
