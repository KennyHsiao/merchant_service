package ordersService

import (
	"com.copo/bo_service/merchant/internal/service/merchantbalanceservice"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	"com.copo/bo_service/merchant/internal/svc"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/neccoys/go-zero-extension/redislock"
	"math"
	"regexp"
	"strconv"
	"time"

	"com.copo/bo_service/common/apimodel/bo"
	"com.copo/bo_service/common/apimodel/bo/searchBO"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/config"
	"com.copo/bo_service/merchant/internal/service/merchantchannelrateservice"
	"com.copo/bo_service/merchant/internal/types"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/gioco-play/gozzle"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type OrdersService struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrdersService(ctx context.Context, svcCtx *svc.ServiceContext) OrdersService {
	return OrdersService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 商户请求参数、處理代付訂單、商户费率验证
func ProxyOrder(db *gorm.DB, req *types.ProxyPayRequestX, ctx context.Context) (rate *types.CorrespondMerChnRate, err error) {
	merchant := &types.Merchant{}
	db.Table("mc_merchants").Where("code = ?", req.MerchantId).Take(merchant)

	req.ProxyPayOrderRequest.OrderAmount = req.OrderAmount.String()
	req.ProxyPayOrderRequest.AccessType = req.AccessType.String()
	// 1. 檢查商户提交參數相關
	err = validateProxyParam(db, req, merchant, ctx)
	if err != nil {
		logx.WithContext(ctx).Errorf("商戶%s,代付提單參數錯誤:%s", merchant.Code, i18n.Sprintf(err.Error()))
		return nil, err
	}

	// 2. 取得商戶配置的費率，以及费率相关验证
	rate, err = checkProxyOrderAndRate(db, merchant, req, ctx)
	if err != nil {
		logx.WithContext(ctx).Errorf("代付提单储存失败%s:%s", err.Error(), i18n.Sprintf(err.Error()))
		return rate, err
	}

	return rate, nil
}

/*
	return:
		1. rate      将该订单指定得费率物件返回上一层
        2. err       將错误返回
*/
func checkProxyOrderAndRate(db *gorm.DB, merchant *types.Merchant, req *types.ProxyPayRequestX, ctx context.Context) (rate2 *types.CorrespondMerChnRate, err error) {

	var orderAmount float64

	orderAmount, _ = strconv.ParseFloat(req.ProxyPayOrderRequest.OrderAmount, 64)

	//检查商户ＡＰＩ提单，代付订单资料是否正确
	err = validProxyPayOrderDataByApi(db, req)
	if err != nil {
		logx.WithContext(ctx).Errorf("检查商户API提单，代付订单资料。%s:%s", err.Error(), i18n.Sprintf(err.Error()))
		return nil, err
	}

	//0. 判断商户是否走智能判断
	// 商戶:
	// 多指定模式。 指定渠道中的渠道，且payTypeSubNo必填
	// 		22/08/24:  多指商戶要能 PayTypeSubNo 沒給能取最小
	// 		22/09/07:  多指商戶要能 PayTypeSubNo 沒給能，走智能判断
	// 单指定模式。 走指定渠道(唯一一个)

	//1. 取得商户对应的代付渠道资料及费率(先收) 计算手续费
	// rate2 := &types.CorrespondMerChnRate{}
	if merchant.BillLadingType == "1" && req.PayTypeSubNo == "" {
		var rateList []types.CorrespondMerChnRate
		rateList, err = merchantchannelrateservice.SmartGetDesignationMerChnRate(db, req.MerchantId, constants.CHN_PAY_TYPE_PROXY_PAY, req.Currency, orderAmount)
		if len(rateList) == 0 {
			return nil, errorz.New(response.MERCHANT_IS_NOT_SETTING_CHANNEL)
		}
		rate2 = &rateList[0]
	} else {
		rate2, err = merchantchannelrateservice.GetDesignationMerChnRate(db, req.MerchantId, constants.CHN_PAY_TYPE_PROXY_PAY, req.Currency, req.PayTypeSubNo, merchant.BillLadingType)
	}

	logx.Infof("渠道资讯及费率的计算类型:%+v", rate2)
	if err != nil {
		logx.WithContext(ctx).Error("商户费率错误。%s:%s", err.Error(), i18n.Sprintf(err.Error()))
		logx.WithContext(ctx).Error("商户模式(提單類型 (0=單指、1=多指)): ", merchant.BillLadingType, "提单交易类型:", constants.CHN_PAY_TYPE_PROXY_PAY)
	}
	if rate2 == nil {
		//未配置渠道，列为失败订单
		logx.WithContext(ctx).Error("商户：{}，提单号：{} 未配置渠道，CorrespondMerChnRate={}", merchant.Code, req.OrderNo, rate2)
		jsonData, _ := json.Marshal(rate2)
		return nil, errorz.New(response.MERCHANT_IS_NOT_SETTING_CHANNEL, string(jsonData))
	} else {
		// 判断提单金额最低金额、最高金额
		if orderAmount < rate2.SingleMinCharge {
			logx.WithContext(ctx).Errorf("付款人:%s,银行账号:%s,%f单笔小于最低代付金额%f", req.DefrayName, req.BankNo, req.OrderAmount, rate2.SingleMinCharge)
			return rate2, errorz.New(response.IS_LESS_MINIMUN_AMOUNT, fmt.Sprintf("%s", req.OrderAmount), fmt.Sprintf("%f", rate2.SingleMinCharge))
		} else if orderAmount > rate2.SingleMaxCharge {
			logx.WithContext(ctx).Errorf("付款人:%s,银行账号:%s,%f单笔大于最高代付金额%f", req.DefrayName, req.BankNo, req.OrderAmount, rate2.SingleMaxCharge)
			return rate2, errorz.New(response.IS_GREATER_MXNIMUN_AMOUNT, fmt.Sprintf("%s", req.OrderAmount), fmt.Sprintf("%f", rate2.SingleMinCharge))
		}

		//  代理-取得商户费率层级编号(需提供merchantCoding、agentLayerNo、payTypeCoding)
		// 2. 补当渠道成本增加时，如果尚未重新配置商户费率，商户费率小于成本时，会退回提单。
		channelPayType := &types.ChannelPayType{}
		db.Table("ch_channel_pay_types").Where("code = ?", rate2.ChannelPayTypesCode).Take(channelPayType)
		if rate2.Fee != 0 {
			if rate2.Fee < channelPayType.Fee && merchant.RateCheck != "0" {
				logx.WithContext(ctx).Errorf("代付提单：%s，商户:%s，代付费率:%f 不可小于渠道成本费率%f ", req.OrderNo, req.MerchantId, rate2.Fee, channelPayType.Fee)
				return rate2, errorz.New(response.RATE_SETTING_ERROR)
			}
		}

		if rate2.HandlingFee != 0 {
			if rate2.HandlingFee < channelPayType.HandlingFee && merchant.RateCheck != "0" {
				logx.WithContext(ctx).Errorf("代付提单：%s，商户:%s，代付费率:%f 不可小于渠道成本费率%f ", req.OrderNo, req.MerchantId, rate2.HandlingFee, channelPayType.HandlingFee)
				return rate2, errorz.New(response.RATE_SETTING_ERROR)
			}
		}
	}

	return rate2, nil
}

/*  代付提單呼叫渠道
	@param respOrder : 代付儲存成功的訂單
    @param rate		 : 商戶配置的費率

	@return error    : call 渠道返回錯誤
*/
func CallChannel_ProxyOrder(context *context.Context, config *config.Config, merReq *types.ProxyPayRequestX, respOrder *types.OrderX, rate *types.CorrespondMerChnRate) (*vo.ProxyPayRespVO, error) {

	span := trace.SpanFromContext(*context)

	precise := utils.GetDecimalPlaces(respOrder.OrderAmount)
	valTrans := strconv.FormatFloat(respOrder.OrderAmount, 'f', precise, 64)

	// 新增请求代付请求app 物件 ProxyPayBO
	ProxyPayBO := bo.ProxyPayBO{
		OrderNo:              respOrder.OrderNo,
		TransactionType:      constants.TRANS_TYPE_PROXY_PAY,
		TransactionAmount:    valTrans,
		ReceiptAccountNumber: respOrder.MerchantBankAccount,
		ReceiptAccountName:   respOrder.MerchantAccountName,
		ReceiptCardProvince:  respOrder.MerchantBankProvince,
		ReceiptCardCity:      respOrder.MerchantBankCity,
		ReceiptCardArea:      "",
		ReceiptCardBranch:    merReq.BranchName,
		ReceiptCardBankCode:  respOrder.MerchantBankNo,
		ReceiptCardBankName:  respOrder.MerchantBankName,
		PlayerId:             merReq.PlayerId,
		Remark:               merReq.Remark,
	}

	// call 渠道app
	ProxyKey, errk := utils.MicroServiceEncrypt(config.ApiKey.ProxyKey, config.ApiKey.PublicKey)
	if errk != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, errk.Error())
	}
	logx.WithContext(*context).Infof("EncryptKey: %s，ProxyKey:%s ，PublicKey:%s ", ProxyKey, config.ApiKey.ProxyKey, config.ApiKey.PublicKey)
	url := fmt.Sprintf("%s:%s/api/proxy-pay", config.Server, rate.ChannelPort)
	chnResp, chnErr := gozzle.Post(url).Timeout(25).Trace(span).Header("authenticationProxykey", ProxyKey).JSON(ProxyPayBO)
	//res, err2 := http.Post(url,"application/json",bytes.NewBuffer(body))
	if chnResp != nil {
		logx.WithContext(*context).Info("response Status:", chnResp.Status())
		logx.WithContext(*context).Info("response Body:", string(chnResp.Body()))
	}

	proxyPayRespVO := &vo.ProxyPayRespVO{}

	if chnErr != nil {
		logx.WithContext(*context).Errorf("渠道返回错误: %s， resp: %#v", chnErr.Error(), chnResp)
		return nil, errorz.New(response.CHANNEL_REPLY_ERROR, chnErr.Error())
	} else if chnResp.Status() != 200 {
		logx.WithContext(*context).Errorf("渠道返回不正确: %d:%s", chnResp.Status(), string(chnResp.Body()))
		return nil, errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("%s", string(chnResp.Body())))
	} else if decodeErr := chnResp.DecodeJSON(proxyPayRespVO); decodeErr != nil {
		logx.WithContext(*context).Errorf("渠道返回错误: %s， resp: %#v", decodeErr.Error(), decodeErr)
		return nil, errorz.New(response.CHANNEL_REPLY_ERROR, decodeErr.Error())
	} else if proxyPayRespVO.Code != "0" {
		return &vo.ProxyPayRespVO{
			Code:    proxyPayRespVO.Code,
			Message: proxyPayRespVO.Message,
		}, nil
	} else if proxyPayRespVO.Data.ChannelOrderNo == "" {
		return &vo.ProxyPayRespVO{
			Code:    proxyPayRespVO.Code,
			Message: proxyPayRespVO.Message,
			Data: vo.ChannelAppProxyResponse{
				ChannelOrderNo: proxyPayRespVO.Data.ChannelOrderNo,
				OrderStatus:    proxyPayRespVO.Data.OrderStatus,
			},
		}, errorz.New(response.INVALID_CHANNEL_ORDER_NO, "ChannelOrderNo: "+proxyPayRespVO.Data.ChannelOrderNo)
	}

	logx.WithContext(*context).Infof("proxyPayRespVO : %#v", proxyPayRespVO)
	return proxyPayRespVO, nil
}

// 目前商戶查單改COPO DB，若有需求再call channel
func CallChannel_ProxyQuery(context *context.Context, config *config.Config, apiUrl string, orderNo string) (*vo.ProxyQueryRespVO, error) {
	span := trace.SpanFromContext(*context)

	proxyQueryBO := &bo.ProxyQueryBO{
		OrderNo: orderNo,
	}

	// call 渠道app
	ProxyKey, errk := utils.MicroServiceEncrypt(config.ApiKey.ProxyKey, config.ApiKey.PublicKey)
	if errk != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, errk.Error())
	}

	chnResp, chnErr := gozzle.Post(apiUrl).Timeout(10).Trace(span).Header("authenticationProxykey", ProxyKey).JSON(proxyQueryBO)
	//res, err2 := http.Post(url,"application/json",bytes.NewBuffer(body))
	logx.WithContext(*context).Info("response Status:", chnResp.Status())
	logx.WithContext(*context).Info("response Body:", string(chnResp.Body()))

	proxyQueryResp := &vo.ProxyQueryRespVO{}

	if chnErr != nil {
		logx.WithContext(*context).Errorf("渠道返回错误: %s， resp: %#v", chnErr.Error(), chnResp)
		return nil, errorz.New(response.CHANNEL_REPLY_ERROR, string(chnResp.Body()))
	} else if chnResp.Status() != 200 {
		logx.WithContext(*context).Errorf("渠道返回不正确: %d", chnResp.Status())
		return nil, errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("%d", chnResp.Status()))
	} else if decodeErr := chnResp.DecodeJSON(proxyQueryResp); decodeErr != nil {
		logx.WithContext(*context).Errorf("渠道返回错误: %s， resp: %#v", decodeErr.Error(), decodeErr)
		return nil, errorz.New(response.CHANNEL_REPLY_ERROR, decodeErr.Error())
	} else if proxyQueryResp.Code != "0" {
		return proxyQueryResp, errorz.New(proxyQueryResp.Code, proxyQueryResp.Message)
	}

	logx.WithContext(*context).Infof("proxyQueryResp : %#v", proxyQueryResp)
	return proxyQueryResp, nil
}

func (l *OrdersService) ChannelCallBackForProxy(req *types.ProxyPayOrderCallBackRequest, orderX *types.OrderX) error {
	redisKey := fmt.Sprintf("%s-%s", orderX.MerchantCode, req.ProxyPayOrderNo)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "proxy-call-back:")
	redisLock.SetExpire(5)
	if isOK, _ := redisLock.Acquire(); isOK {
		if err := l.internalChannelCallBackForProxy(req, orderX); err != nil {
			return err
		}
		defer redisLock.Release()
	} else {
		logx.WithContext(l.ctx).Infof(i18n.Sprintf(response.TRANSACTION_PROCESSING))
		return errorz.New(response.TRANSACTION_PROCESSING)
	}
	return nil
}

/*
	渠道(代付)回调用，更新状态用(Service)
	目前仅提供 1.渠道回调  TODO 2.scheduled回调使用
	提单order_status为[1=成功][2=失败]时，不接受回调的变更
*/
func (l *OrdersService) internalChannelCallBackForProxy(req *types.ProxyPayOrderCallBackRequest, orderX *types.OrderX) error {

	//订单状态为[成功]或[失败]，判定为已结单，以及[人工處理單]不接受回調
	if req.ChannelResultStatus == "1" && orderX.Status == constants.SUCCESS || req.ChannelResultStatus == "1" && orderX.Status == constants.FAIL {
		logx.WithContext(l.ctx).Infof("代付订单：%s，提单状态为：%s，判定为已结单，直接返回。", orderX.OrderNo, orderX.Status)
		return nil
	}
	if orderX.Status == constants.SUCCESS || orderX.Status == constants.FAIL || orderX.Status == constants.FROZEN {
		logx.WithContext(l.ctx).Infof("代付订单：%s，提单状态为：%s，判定为已结单，不接受回调变更。", orderX.OrderNo, orderX.Status)
		//TODO 是否有需要回寫代付歷程?
		return errorz.New(response.PROXY_PAY_IS_CLOSE, "此提单目前已为结单状态")
	} else if orderX.PersonProcessStatus != constants.PERSON_PROCESS_STATUS_NO_ROCESSING {
		logx.WithContext(l.ctx).Infof("代付订单：%s，人工处里状态为：%s，判定已进入人工处里阶段，不接受回调变更。", orderX.OrderNo, orderX.PersonProcessStatus)
		//TODO 是否有需要回寫代付歷程?
		return errorz.New(response.PROXY_PAY_IS_PERSON_PROCESS, "提单目前为人工处里阶段，不可回调变更")
	} else if req.ChannelResultStatus == constants.CALL_BACK_STATUS_SUCCESS && req.Amount != orderX.OrderAmount {
		//代付回调金额若不等于订单金额，订单转 人工处理，并塞error_message
		logx.WithContext(l.ctx).Errorf("代付回调金额不等于订单金额，订单转人工处理。单号:%s", orderX.OrderNo)
		orderX.ErrorNote = "代付渠道回调金额不等于订单金额"
		orderX.ErrorType = "1"
		orderX.PersonProcessStatus = constants.PERSON_PROCESS_WAIT_CHANGE
		orderX.RepaymentStatus = constants.REPAYMENT_WAIT

		if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(orderX).Error; errUpdate != nil {
			logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
		}

		return errorz.New(response.PROXY_PAY_AMOUNT_INVALID, "代付回调金额与订单金额不符")
	}
	/*更新訂單:
	1. 訂單狀態(依渠道回調決定)
	2. 還款狀態(預設[0]：不需还款)，若渠道回調失敗單，則[1]：待还款
	*/
	//订单预设还款状态为"不需还款"，更新為待还款
	orderX.Status = getOrderStatus(req.ChannelResultStatus)
	orderX.RepaymentStatus = constants.REPAYMENT_NOT //还款状态：([0]：不需还款、1：待还款、2：还款成功、3：还款失败)，预设不需还款
	if req.ChannelResultStatus == constants.CALL_BACK_STATUS_FAIL {
		orderX.RepaymentStatus = constants.REPAYMENT_WAIT //还款状态：(0：不需还款、[1]：待还款、2：还款成功、3：还款失败)，预设不需还款
		if req.ChannelResultNote != "" {
			orderX.ErrorNote = "渠道回调:-" + req.ChannelResultNote //失败时，写入渠道返还的讯息
		} else {
			orderX.ErrorNote = "渠道回调: 交易失败"
		}
	}

	orderX.UpdatedBy = req.UpdatedBy
	orderX.UpdatedAt = time.Now().UTC()
	orderX.CallBackStatus = req.ChannelResultStatus
	orderX.ChannelCallBackAt = time.Now().UTC()
	orderX.TransAt = types.JsonTime{}.New()

	if orderX.ChannelOrderNo == "" {
		orderX.ChannelOrderNo = req.ChannelOrderNo
	}
	if req.ChannelCharge != 0 {
		//渠道有回傳渠道手續費
	}

	// 更新订单
	if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(orderX).Error; errUpdate != nil {
		logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
	}

	// 更新訂單訂單歷程 (不抱錯)
	if orderX.Status == constants.SUCCESS {
		if err4 := l.svcCtx.MyDB.Table("tx_order_actions").Create(&types.OrderActionX{
			OrderAction: types.OrderAction{
				OrderNo:     orderX.OrderNo,
				Action:      "SUCCESS",
				UserAccount: orderX.MerchantCode,
				Comment:     "",
			},
		}).Error; err4 != nil {
			logx.Error("紀錄訂單歷程出錯:%s", err4.Error())
		}
	}

	var errRpc error
	if req.ChannelResultStatus == constants.CALL_BACK_STATUS_SUCCESS {

	} else if req.ChannelResultStatus == constants.CALL_BACK_STATUS_FAIL { //==========渠道回調失敗=========START
		logx.WithContext(l.ctx).Info("代付订单回调状态为[失败]，开始还款=======================================>", orderX.Order.OrderNo)
		//呼叫RPC
		balanceType, errBalance := merchantbalanceservice.GetBalanceTypeByOrder(l.svcCtx.MyDB, req.ProxyPayOrderNo)
		if errBalance != nil {
			return errBalance
		}

		//當訂單還款狀態為待还款
		if orderX.RepaymentStatus == constants.REPAYMENT_WAIT {
			//将商户钱包加回 (merchantCode, orderNO)，更新狀態為失敗單
			var resRpc *transaction.ProxyPayFailResponse
			if balanceType == "DFB" {
				resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTransactionFail_DFB(l.ctx, &transaction.ProxyPayFailRequest{
					MerchantCode: orderX.MerchantCode,
					OrderNo:      orderX.OrderNo,
				})
			} else if balanceType == "XFB" {
				resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTransactionFail_XFB(l.ctx, &transaction.ProxyPayFailRequest{
					MerchantCode: orderX.MerchantCode,
					OrderNo:      orderX.OrderNo,
				})
			}

			if errRpc != nil {
				logx.WithContext(l.ctx).Errorf("代付提单回调 %s 还款失败。 Err: %s", orderX.OrderNo, errRpc.Error())
				orderX.RepaymentStatus = constants.REPAYMENT_FAIL
				return errorz.New(response.FAIL, errRpc.Error())
			} else {
				logx.WithContext(l.ctx).Infof("代付還款rpc完成，%s 錢包還款完成: %#v", balanceType, resRpc)
				orderX.RepaymentStatus = constants.REPAYMENT_SUCCESS
				//TODO 收支紀錄
			}
		}
	} //  ==========渠道回調失敗=========END

	//TODO 是否有需還款成功才回調給商戶?
	//若訂單單來源為API且尚未回調給商戶，進行回調給商戶
	if orderX.IsMerchantCallback == constants.IS_MERCHANT_CALLBACK_NO && (orderX.Status == constants.SUCCESS || orderX.Status == constants.FAIL) {
		if req.ChannelResultStatus == constants.CALL_BACK_STATUS_SUCCESS {
			logx.WithContext(l.ctx).Infof("代付订单回调状态为[成功]，主动回调商户API订单：%s=======================================>", orderX.Order.OrderNo)
		} else if req.ChannelResultStatus == constants.CALL_BACK_STATUS_FAIL {
			logx.WithContext(l.ctx).Infof("代付订单回调状态为[失敗]，主动回调商户API订单：%s=======================================>", orderX.OrderNo)
		}

		go func() {
			if errPoseMer := merchantsService.PostCallbackToMerchant(l.svcCtx.MyDB, &l.ctx, orderX); errPoseMer != nil {
				//不拋錯
				logx.WithContext(l.ctx).Error("回調商戶錯誤:", errPoseMer)

			} else {
				//更改回調商戶状态
				if orderX.IsMerchantCallback == constants.MERCHANT_CALL_BACK_NO {
					orderX.IsMerchantCallback = constants.MERCHANT_CALL_BACK_YES
					orderX.MerchantCallBackAt = time.Now().UTC()
				}
				// 更新订单
				if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(orderX).Error; errUpdate != nil {
					logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
				}
			}
		}()

	}

	// 更新订单
	if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(orderX).Error; errUpdate != nil {
		logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
	}

	return nil
}

/*
	TODO 内部代付业务，还款用API(必须是 1.待还款、3.还款失败的提单才可更新)
*/
func ProxyPayRepayment() {

}

func getOrderStatus(channelResultStatus string) string {

	var orderStatus string
	switch channelResultStatus { //渠道回調狀態(0:處理中1:成功2:失敗)
	case "0": // 代處理
		orderStatus = constants.TRANSACTION
	case "1": // 成功
		orderStatus = constants.SUCCESS
	case "2": // 失敗
		orderStatus = constants.FAIL
	default:
		orderStatus = constants.TRANSACTION
	}
	return orderStatus
}

func getMerAllLayerFeeRateInfo(db *gorm.DB, searchBo *searchBO.MerchantLayerRateDataSearchBO) {

}

func validProxyPayOrderDataByApi(db *gorm.DB, req *types.ProxyPayRequestX) (err error) {
	var txOrder types.OrderX
	//1.检查订单号是否重复
	orderErr := db.Table("tx_orders").
		Select("merchant_code, merchant_order_no").
		Where("merchant_code = ?", req.MerchantId).
		Where("merchant_order_no = ?", req.OrderNo).Take(&txOrder).Error

	if orderErr == nil {
		return errorz.New(response.REPEAT_ORDER_NO, "Merchant OrderNo: "+req.OrderNo)
	} else if orderErr != nil && !errors.Is(orderErr, gorm.ErrRecordNotFound) {
		return errorz.New(response.SYSTEM_ERROR, err.Error())
	}

	//2.验证币别是否可使用
	isCheck := checkCurrencyCodeByApi(db, req.MerchantId, req.Currency)
	if !isCheck {
		return errorz.New(response.CURRENCY_INCONSISTENT, "currency: "+req.Currency)
	}
	return nil
}

func checkCurrencyCodeByApi(db *gorm.DB, merchantCode string, currency string) bool {
	MerchantCurrency := &types.MerchantCurrency{}
	if err := db.Table("mc_merchant_currencies").
		Where("merchant_code = ? AND currency_code = ?", merchantCode, currency).
		Find(&MerchantCurrency).Error; err != nil {
		return false
	}
	return true
}

func autoFillBankName(db *gorm.DB, req *types.ProxyPayRequestX) (err error) {
	bank := &types.Bank{}
	if req.BankId == "" || req.BankId == "000" { //银行代码给000 省略銀行代碼
		return nil
	} else {
		if err = db.Table("bk_banks").Where("bank_no", req.BankId).Where("currency_code", req.Currency).Take(bank).Error; err != nil {
			return errorz.New(response.BANK_CODE_INVALID, err.Error(), req.BankId)
		}
		req.BankName = bank.BankName
		return nil
	}
}

func validateProxyParam(db *gorm.DB, req *types.ProxyPayRequestX, merchant *types.Merchant, ctx context.Context) (err error) {

	var orderAmount float64
	orderAmount, _ = strconv.ParseFloat(req.ProxyPayOrderRequest.OrderAmount, 64)

	// 檢查簽名
	checkSign := utils.VerifySign(req.Sign, req.ProxyPayOrderRequest, merchant.ScrectKey, ctx)
	if !checkSign {
		return errorz.New(response.INVALID_SIGN)
	}
	// 檢查新增USDT 钱包地址判断 协定固定 USDT-TRC20
	//if req.Currency == "USDT" {
	//	if isMatch, _ := regexp.MatchString(constants.REGEXP_WALLET_TRC, req.BankNo); !isMatch {
	//		return errorz.New(response.INVALID_USDT_WALLET_ADDRESS, "USDT_WALLET_ADDRESS: "+req.BankNo)
	//	}
	//}

	//======業務參數驗證==========
	if req.AccessType == "" || req.AccessType != constants.ACCESS_TYPE_PROXY {
		return errorz.New(response.INVALID_ACCESS_TYPE, "AccessType: "+req.AccessType.String())
	}
	if req.MerchantId == "" {
		return errorz.New(response.INVALID_MERCHANT_CODE, "MerchantId: "+req.MerchantId)
	}
	if req.OrderNo == "" {
		return errorz.New(response.INVALID_ORDER_NO, "OrderNo: "+req.OrderNo)
	}

	//4.验证开户行号(银行代码)(必填)(格式必须都为数字)(长度只能为3码)
	//isMatch, _ := regexp.MatchString(constants.REGEXP_BANK_ID, req.BankId)
	//if req.BankId == "" || !isMatch || len(req.BankId) != 3 {
	//	logx.WithContext(ctx).Error("开户行号格式不符: ", req.BankId)
	//	return errorz.New(response.INVALID_BANK_ID, "BankId: "+req.BankId)
	//}
	//5.验证开户行名(必填)
	if req.BankName == "" {
		return errorz.New(response.INVALID_BANK_NAME, "BankName: "+req.BankName)
	}

	//6.验证银行卡号(必填)(必须为数字)(长度必须在13~22码)
	isMatch2, _ := regexp.MatchString(constants.REGEXP_BANK_ID, req.BankNo)
	currencyCode := req.Currency
	if currencyCode == constants.CURRENCY_THB {
		if req.BankNo == "" || len(req.BankNo) < 10 || len(req.BankNo) > 16 || !isMatch2 {
			logx.WithContext(ctx).Error("銀行卡號檢查錯誤，需10-16碼內：", req.BankNo)
			return errorz.New(response.INVALID_BANK_NO, "BankNo: "+req.BankNo)
		}
	} else if currencyCode == constants.CURRENCY_CNY {
		if req.BankNo == "" || len(req.BankNo) < 13 || len(req.BankNo) > 22 || !isMatch2 {
			logx.WithContext(ctx).Error("銀行卡號檢查錯誤，需13-22碼內：", req.BankNo)
			return errorz.New(response.INVALID_BANK_NO, "BankNo: "+req.BankNo)
		}
	}

	//7.验证开户人姓名(必填)
	if req.DefrayName == "" {
		return errorz.New(response.INVALID_DEFRAY_NAME, "DefrayName: "+req.DefrayName)
	}

	//8.验证交易金额(必填)
	if orderAmount <= 0 {
		logx.WithContext(ctx).Error("金额错误", req.ProxyPayOrderRequest.OrderAmount)
		return errorz.New(response.INVALID_AMOUNT, "OrderAmount: "+fmt.Sprintln("%d", req.OrderAmount))
	} else if req.Currency == "CNY" && math.Mod(orderAmount, 1) != 0 {
		return errorz.New(response.INVALID_AMOUNT, "提交金额不为小数。orderAmount: "+req.ProxyPayOrderRequest.OrderAmount)
	}

	isMatch3, _ := regexp.MatchString(constants.REGEXP_URL, req.NotifyUrl)
	//9.验证回调URL格式
	if req.NotifyUrl == "" || !isMatch3 {
		return errorz.New(response.INVALID_NOTIFY_URL, "NotifyUrl: "+req.NotifyUrl)
	}

	//10.验证语系(目前仅支援简体中文)
	//if req.Language == "" || req.Language != constants.LANGUAGE_ZH_CN {
	//	return errorz.New(response.INVALID_LANGUAGE_CODE, "Language: "+req.Language)
	//}

	//11 判斷銀行代號自動填入名稱
	if err = autoFillBankName(db, req); err != nil {
		logx.WithContext(ctx).Error("银行代码错误: ", err.Error())
		return errorz.New(response.INVALID_BANK_ID, "BankID: "+req.BankId)
	}

	return nil
}
