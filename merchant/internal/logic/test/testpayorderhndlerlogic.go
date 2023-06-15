package test

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/logic/payorder"
	"com.copo/bo_service/merchant/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestPayOrderHndlerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestPayOrderHndlerLogic(ctx context.Context, svcCtx *svc.ServiceContext) TestPayOrderHndlerLogic {
	return TestPayOrderHndlerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestPayOrderHndlerLogic) TestPayOrderHndler(req *types.TestPayOrderRequest) (resp *types.PayOrderResponse, err error) {
	var payOrderReq types.PayOrderRequestX
	var merchantChannelRate types.MerchantChannelRate
	var merchant types.Merchant

	merchantId := "ME00015"
	copier.Copy(&payOrderReq, &req)

	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", merchantId).
		Take(&merchant).Error; err != nil {
		return nil, errorz.New(response.MERCHANT_IS_NOT_SETTLE)
	}
	screctKey := merchant.ScrectKey

	if err = l.svcCtx.MyDB.Table("mc_merchant_channel_rate").
		Where("merchant_code = ?", merchantId).
		Where("channel_pay_types_code = ?", req.ChannelCode+req.PayType).
		Take(&merchantChannelRate).Error; err != nil {
		return nil, errorz.New(response.INVALID_MERCHANT_OR_CHANNEL_PAYTYPE, fmt.Sprintf("商户代码[%s]或支付类型代码[%s]或幣別[%s]错误或指定渠道设定错误或关闭或维护", merchantId, req.PayType, req.Currency))
	}

	payOrderReq.PayOrderRequest.AccessType = "1"
	payOrderReq.NotifyUrl = "https://merchant.copo.vip/dior/merchant-api/test_merchant_pay-call-back"
	payOrderReq.PageUrl = "https://docs.goldenf.vip/"
	payOrderReq.PageFailedUrl = "https://docs.goldenf.vip/"
	payOrderReq.Language = "ZH-CN"
	payOrderReq.OrderNo = model.GenerateOrderNo("TEST")
	payOrderReq.OrderName = "TEST"
	payOrderReq.UserId = "测试员"
	payOrderReq.MerchantId = merchantId
	payOrderReq.PayOrderRequest.OrderAmount = payOrderReq.OrderAmount.String()
	payOrderReq.PayOrderRequest.AccessType = payOrderReq.AccessType.String()
	payOrderReq.PayTypeNo = json.Number(merchantChannelRate.DesignationNo)
	payOrderReq.PayOrderRequest.PayTypeNo = merchantChannelRate.DesignationNo
	payOrderReq.PayOrderRequest.PlayerId = "COPO_TEST"
	payOrderReq.Phone = "123456"
	payOrderReq.Address = "Address"
	payOrderReq.City = "City"
	payOrderReq.ZipCode = "3365"
	payOrderReq.Country = "IN"
	payOrderReq.Email = "aaa@aa.com"
	payOrderReq.Sign = utils.SortAndSign2(payOrderReq.PayOrderRequest, screctKey)
	payOrderReq.MyIp = req.IP
	payOrderReq.BankCode = req.BankCode

	pl := payorder.NewPayOrderLogic(l.ctx, l.svcCtx)
	return pl.PayOrder(payOrderReq)
}
