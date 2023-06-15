package test

import (
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/logic/test"
	"com.copo/bo_service/merchant/internal/svc"
	"net/http"
)

func HealthCheckHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := test.NewHealthCheckLogic(r.Context(), ctx)
		resp, err := l.HealthCheck()
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
