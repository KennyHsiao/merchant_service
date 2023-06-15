package payorder

import (
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestMerchantPayCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestMerchantPayCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) TestMerchantPayCallBackLogic {
	return TestMerchantPayCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestMerchantPayCallBackLogic) TestMerchantPayCallBack(req *types.TestMerchantPayCallBackRequest) (resp string, err error) {

	logx.WithContext(l.ctx).Infof("測試回調接收 %#v", req)
	resp = "success"

	return
}
