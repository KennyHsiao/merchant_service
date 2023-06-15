package payorder

import (
	"com.copo/bo_service/common/constants/redisKey"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"encoding/json"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayOrderChannelRedirectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayOrderChannelRedirectLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayOrderChannelRedirectLogic {
	return PayOrderChannelRedirectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayOrderChannelRedirectLogic) PayOrderChannelRedirect(req types.PayOrderChannelBankRequest) (resp *types.PayOrderChannelRedirectResponse, err error) {
	// 取得存在Redis的資料

	redisKey := redisKey.CACHE_PAY_ORDER_CHANNEL_REDIRECT + req.OrderNo
	logx.WithContext(l.ctx).Infof("redisKey : ", redisKey)
	result, err := l.svcCtx.RedisClient.Get(l.ctx, redisKey).Result()
	if err != nil {
		return nil, errorz.New(response.INVALID_ORDER_NO, err.Error())
	}
	expiration, err := l.svcCtx.RedisClient.TTL(l.ctx, redisKey).Result()
	if err != nil {
		return nil, errorz.New(response.INVALID_ORDER_NO, err.Error())
	}
	if err = json.Unmarshal([]byte(result), &resp); err != nil {
		return nil, errorz.New(response.API_PARAMETER_TYPE_ERROE, err.Error())
	}
	resp.Expiration = expiration.Seconds()

	return
}
