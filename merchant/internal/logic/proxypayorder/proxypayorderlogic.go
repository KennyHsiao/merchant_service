package proxypayorder

import (
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/model"
	"com.copo/bo_service/merchant/internal/service/autoFillBankIdService"
	"com.copo/bo_service/merchant/internal/service/merchantbalanceservice"
	"com.copo/bo_service/merchant/internal/service/merchantchannelrateservice"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	ordersService "com.copo/bo_service/merchant/internal/service/orders"
	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"errors"
	"fmt"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/jinzhu/copier"
	"github.com/neccoys/go-zero-extension/redislock"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"regexp"
	"strconv"
	"strings"
)

type ProxyPayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyPayOrderLogic {
	return ProxyPayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyPayOrderLogic) ProxyPayOrder(merReq *types.ProxyPayRequestX) (*types.ProxyPayOrderResponse, error) {
	var resp *types.ProxyPayOrderResponse
	var err error
	redisKey := fmt.Sprintf("%s-%s", merReq.MerchantId, merReq.OrderNo)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "proxy-pay:")
	redisLock.SetExpire(5)
	if isOK, redisErr := redisLock.Acquire(); redisErr != nil {
		logx.WithContext(l.ctx).Errorf("redisLock error: %s", redisErr.Error())
		return nil, errorz.New(response.TRANSACTION_PROCESSING)
	} else if isOK {
		if resp, err = l.internalProxyPayOrder(merReq); err != nil {
			return nil, err
		}
		defer redisLock.Release()
	} else {
		return nil, errorz.New(response.TRANSACTION_PROCESSING)
	}
	return resp, err
}

func (l *ProxyPayOrderLogic) internalProxyPayOrder(merReq *types.ProxyPayRequestX) (*types.ProxyPayOrderResponse, error) {
	logx.WithContext(l.ctx).Infof("Enter proxy-order: %#v", merReq)
	resp := &types.ProxyPayOrderResponse{}
	// 1. 檢查白名單及商户号
	merchant, errWhite := l.CheckMerAndWhiteList(merReq)
	if errWhite != nil {
		logx.WithContext(l.ctx).Error("商戶號及白名單檢查錯誤: ", errWhite.Error())
		resp.RespCode = errWhite.Error()
		resp.RespMsg = i18n.Sprintf(errWhite.Error())
		return resp, errWhite
	}

	// 2. 處理商户提交参数、訂單驗證，並返回商戶費率
	rate, errCreate := ordersService.ProxyOrder(l.svcCtx.MyDB, merReq, l.ctx)
	if errCreate != nil {
		logx.WithContext(l.ctx).Errorf("代付提單商户提交参数驗證錯誤: %s:%s", errCreate.Error(), i18n.Sprintf(errCreate.Error()))
		resp.RespCode = errCreate.Error()
		resp.RespMsg = i18n.Sprintf(errCreate.Error())
		return resp, errCreate
	}

	// 22/08/31  代付提單沒給bankId, 用銀行名稱判斷補給 (目前僅支援MLB代付)
	if merReq.BankId == "" && (rate.ChannelCode == "CHN000152" || rate.ChannelCode == "CHN000119" || rate.ChannelCode == "CHN000001") {
		if err := autoFillBankIdService.AutoFillBankId(l.ctx, l.svcCtx.MyDB, merReq); err != nil {
			logx.WithContext(l.ctx).Errorf("银行名称转代码错误 BankName:%s: %s", merReq.BankName, err.Error())
			return nil, errorz.New(response.INVALID_BANK_ID, "BankName: "+merReq.BankName)
		}
	} else {
		//验证开户行号(银行代码)(必填)(格式必须都为数字)(长度只能为3码)
		isMatch, _ := regexp.MatchString(constants.REGEXP_BANK_ID, merReq.BankId)
		if merReq.BankId == "" || !isMatch || len(merReq.BankId) != 3 {
			logx.WithContext(l.ctx).Error("开户行号格式不符: ", merReq.BankId)
			return nil, errorz.New(response.INVALID_BANK_ID, "BankId: "+merReq.BankId)
		}
	}

	balanceType, errBalance := merchantbalanceservice.GetBalanceType(l.svcCtx.MyDB, rate.ChannelCode, constants.ORDER_TYPE_DF)
	if errBalance != nil {

		resp.RespCode = errBalance.Error()
		resp.RespMsg = i18n.Sprintf(errBalance.Error())
		return resp, errBalance
	}

	// 3. 依balanceType决定要呼叫哪种transaction方法
	//    呼叫 transaction rpc (merReq, rate) (ProxyOrderNo) 并产生订单

	//產生rpc 代付需要的請求的資料物件
	ProxyPayOrderRequest, rateRpc := generateRpcdata(merReq, rate)
	logx.WithContext(l.ctx).Infof("call transactionclient")

	var errRpc error
	var res *transactionclient.ProxyOrderResponse
	if balanceType == "DFB" {
		res, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTranaction_DFB(l.ctx, &transaction.ProxyOrderRequest{
			Req:  ProxyPayOrderRequest,
			Rate: rateRpc,
		})
	} else if balanceType == "XFB" {
		res, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTranaction_XFB(l.ctx, &transaction.ProxyOrderRequest{
			Req:  ProxyPayOrderRequest,
			Rate: rateRpc,
		})
	}
	if errRpc != nil {
		logx.WithContext(l.ctx).Error("代付提單Rpc呼叫失败:", errRpc.Error())
		resp.RespCode = response.FAIL
		return nil, errRpc
	} else if res.Code != "000" {
		logx.WithContext(l.ctx).Infof("代付交易rpc返回失败，%s 錢包扣款失败: %#v", balanceType, res)
		resp.RespCode = res.Code
		resp.RespMsg = res.Message
		return resp, nil
	} else {
		logx.WithContext(l.ctx).Infof("代付交易rpc完成，%s 錢包扣款完成: %#v", balanceType, res)
	}

	var queryErr error
	var respOrder = &types.OrderX{}
	if respOrder, queryErr = model.QueryOrderByOrderNo(l.svcCtx.MyDB, res.ProxyOrderNo, ""); queryErr != nil {
		logx.WithContext(l.ctx).Errorf("撈取代付訂單錯誤: %s, respOrder:%#v", queryErr, respOrder)
		return nil, errorz.New(response.FAIL, queryErr.Error())
	}

	// 4: call channel (不論是否有成功打到渠道，都要返回給商戶success，一渠道返回訂單狀態決定此訂單狀態(代處理/處理中))
	var errCHN error
	var proxyPayRespVO *vo.ProxyPayRespVO
	proxyPayRespVO, errCHN = ordersService.CallChannel_ProxyOrder(&l.ctx, &l.svcCtx.Config, merReq, respOrder, rate)
	logx.WithContext(l.ctx).Infof("渠道返回: %v", proxyPayRespVO)

	//5. 返回給商戶物件
	var proxyResp types.ProxyPayOrderResponse
	i18n.SetLang(language.English)

	// ******************20220920 新增智能判斷 START******************
	if proxyPayRespVO.Code == "202" && merReq.PayTypeSubNo == "" && merchant.BillLadingType == "1" { //多指定模式，payTypeSubNo 沒給值，Call Channel 餘額不足，系統智能判斷決定代付渠道
		/**
		1. 加回錢包。
		   loop
		2. 重新取得商户对应的代付渠道资料及费率 计算手续费。
		3. Call Channel
		   loop
		*/
		proxyPayRespVO, errCHN = l.smartProxyPayProcess(merReq, ProxyPayOrderRequest, respOrder, &balanceType)

	}
	//******************20220920 新增智能判斷 END******************

	if respOrder, queryErr = model.QueryOrderByOrderNo(l.svcCtx.MyDB, res.ProxyOrderNo, ""); queryErr != nil {
		logx.WithContext(l.ctx).Errorf("撈取代付訂單錯誤: %s, respOrder:%#v", queryErr, respOrder)
		return nil, errorz.New(response.FAIL, queryErr.Error())
	}

	if errCHN != nil || proxyPayRespVO.Code != "0" {
		logx.WithContext(l.ctx).Errorf("代付提單: %s ，渠道返回: %v， Err:%s", respOrder.OrderNo, proxyPayRespVO, errCHN)

		//若call渠道timeout则保持订单状态[待处理]，也不还款，排程去反查
		if strings.Index(proxyPayRespVO.Message, "context deadline exceeded") > -1 {
			respOrder.Memo = "请求渠道超时: " + proxyPayRespVO.Message

			// 更新订单
			if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
				logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
			}
		} else {
			//将商户钱包加回 (merchantCode, orderNO)，更新狀態為失敗單
			var resRpc *transaction.ProxyPayFailResponse
			ctx := context.Background()
			logx.WithContext(l.ctx).Infof("还款Transaction: %s", ctx)
			if proxyPayRespVO.Code != response.DEDUCTION_FAILED { //智能订单判断还款失败
				if balanceType == "DFB" {
					resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTransactionFail_DFB(ctx, &transaction.ProxyPayFailRequest{
						MerchantCode: respOrder.MerchantCode,
						OrderNo:      respOrder.OrderNo,
					})
				} else if balanceType == "XFB" {
					resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTransactionFail_XFB(ctx, &transaction.ProxyPayFailRequest{
						MerchantCode: respOrder.MerchantCode,
						OrderNo:      respOrder.OrderNo,
					})
				}
			}

			//因在transaction_service 已更新過訂單，重新抓取訂單
			if respOrder, queryErr = model.QueryOrderByOrderNo(l.svcCtx.MyDB, res.ProxyOrderNo, ""); queryErr != nil {
				logx.WithContext(l.ctx).Errorf("撈取代付訂單錯誤: %s, respOrder:%#v", queryErr, respOrder)
				return nil, errorz.New(response.FAIL, queryErr.Error())
			}

			//處理渠道回傳錯誤訊息
			proxyResp.RespCode = response.CHANNEL_REPLY_ERROR
			if errCHN != nil {
				proxyResp.RespMsg = i18n.Sprintf(response.CHANNEL_REPLY_ERROR) + ": Error: " + i18n.Sprintf(errCHN.Error())
				respOrder.ErrorType = "2" //   1.渠道返回错误	2.渠道异常	3.商户参数错误	4.账户为黑名单	5.其他
				respOrder.ErrorNote = "渠道异常: " + i18n.Sprintf(errCHN.Error())
				respOrder.Status = constants.FAIL
				// 更新订单
				if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
					logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
				}

			} else if proxyPayRespVO.Code != "0" {
				respOrder.ErrorType = "1" //   1.渠道返回错误	2.渠道异常	3.商户参数错误	4.账户为黑名单	5.其他
				respOrder.ErrorNote = "Code:" + proxyPayRespVO.Code + " Message: " + proxyPayRespVO.Message
				respOrder.Status = constants.FAIL
				proxyResp.RespMsg = i18n.Sprintf(response.CHANNEL_REPLY_ERROR) + ": Code: " + proxyPayRespVO.Code + " Message: " + proxyPayRespVO.Message

				// 更新订单
				if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
					logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
				}
			}

			if errRpc != nil {
				logx.WithContext(l.ctx).Errorf("代付提单 %s 还款失败。 Err: %s", respOrder.OrderNo, errRpc.Error())
				respOrder.RepaymentStatus = constants.REPAYMENT_FAIL

				// 更新订单
				if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
					logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
				}

				return nil, errorz.New(response.FAIL, errRpc.Error())
			} else {
				logx.WithContext(l.ctx).Infof("代付還款rpc完成，%s 錢包還款完成: %#v", balanceType, resRpc)
				respOrder.RepaymentStatus = constants.REPAYMENT_SUCCESS
			}

			// 更新订单
			if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
				logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
			}
		}

	} else {

		respOrder.ChannelOrderNo = proxyPayRespVO.Data.ChannelOrderNo
		//条整订单状态从"待处理" 到 "交易中"
		respOrder.Status = constants.TRANSACTION
		proxyResp.RespCode = response.API_SUCCESS
		proxyResp.RespMsg = i18n.Sprintf(response.API_SUCCESS) //固定回商戶成功

		// 更新订单
		if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
			logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
		}
	}

	// 依渠道返回给予订单状态
	var orderStatus string
	if respOrder.Status == constants.FAIL {
		orderStatus = "2"
	} else {
		orderStatus = "0"
	}

	proxyResp.MerchantId = respOrder.MerchantCode
	proxyResp.OrderNo = respOrder.MerchantOrderNo
	proxyResp.PayOrderNo = respOrder.OrderNo
	proxyResp.OrderStatus = orderStatus //渠道返回成功: "處理中" 失敗: "失敗"
	proxyResp.Sign = utils.SortAndSign2(proxyResp, merchant.ScrectKey)

	return &proxyResp, nil
}

//检查商户号是否存在以及IP是否为白名单，若无误则返回"商户物件"
func (l *ProxyPayOrderLogic) CheckMerAndWhiteList(req *types.ProxyPayRequestX) (merchant *types.Merchant, err error) {

	// 檢查白名單
	if err = l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", req.MerchantId).Where("status = ?", "1").Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.DATA_NOT_FOUND, err.Error())
		} else if err == nil && merchant != nil && merchant.Status != constants.MerchantStatusEnable {
			return nil, errorz.New(response.MERCHANT_ACCOUNT_NOT_FOUND, "商户号:"+merchant.Code)
		} else {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	if isWhite := merchantsService.IPChecker(req.Ip, merchant.ApiIP); !isWhite {
		return nil, errorz.New(response.API_IP_DENIED, "IP: "+req.Ip)
	}

	merchantCurrency := &types.MerchantCurrency{}
	if errCurrency := l.svcCtx.MyDB.Table("mc_merchant_currencies").
		Where("merchant_code = ? AND currency_code = ? AND status = ?", req.MerchantId, req.Currency, "1").
		Take(merchantCurrency).Error; errCurrency != nil {
		return nil, errorz.New(response.MERCHANT_CURRENCY_NOT_SET)
	} else if req.PayTypeSubNo == "" && merchant.BillLadingType == "1" && merchantCurrency.IsDisplayPtBalance == "1" {
		// 23/05/16: 商戶多指定模式，沒給PayTypeSubNo，且開啟子錢包模式報錯。
		return nil, errorz.New(response.MERCHANT_NOT_SUPPORT_SUB_WALLET)
	}

	//TODO 檢查幣別啟用、商戶錢包是否啟用禁運
	return merchant, nil
}

// 產生rpc 代付需要的請求的資料物件
func generateRpcdata(merReq *types.ProxyPayRequestX, rate *types.CorrespondMerChnRate) (*transaction.ProxyPayOrderRequest, *transaction.CorrespondMerChnRate) {
	//var orderAmount float64
	//if s, ok := merReq.OrderAmount.(string); ok {
	//	if s, err := strconv.ParseFloat(s, 64); err == nil {
	//		orderAmount = s
	//	}
	//} else if f, ok := merReq.OrderAmount.(float64); ok {
	//	orderAmount = f
	//}
	orderAmount, _ := strconv.ParseFloat(merReq.ProxyPayOrderRequest.OrderAmount, 64)
	ProxyPayOrderRequest := &transaction.ProxyPayOrderRequest{
		AccessType:   merReq.AccessType.String(),
		MerchantId:   merReq.MerchantId,
		Sign:         merReq.Sign,
		NotifyUrl:    merReq.NotifyUrl,
		Language:     merReq.Language,
		OrderNo:      merReq.OrderNo,
		BankId:       merReq.BankId,
		BankName:     merReq.BankName,
		BankProvince: merReq.BankProvince,
		BankCity:     merReq.BankCity,
		BranchName:   merReq.BranchName,
		BankNo:       merReq.BankNo,
		OrderAmount:  orderAmount,
		DefrayName:   merReq.DefrayName,
		DefrayId:     merReq.DefrayId,
		DefrayMobile: merReq.DefrayMobile,
		DefrayEmail:  merReq.DefrayEmail,
		Currency:     merReq.Currency,
		PayTypeSubNo: merReq.PayTypeSubNo,
	}
	rateRpc := &transaction.CorrespondMerChnRate{
		MerchantCode:        rate.MerchantCode,
		ChannelPayTypesCode: rate.ChannelPayTypesCode,
		ChannelCode:         rate.ChannelCode,
		PayTypeCode:         rate.PayTypeCode,
		Designation:         rate.Designation,
		DesignationNo:       rate.DesignationNo,
		Fee:                 rate.Fee,
		HandlingFee:         rate.HandlingFee,
		ChFee:               rate.ChFee,
		ChHandlingFee:       rate.ChHandlingFee,
		SingleMinCharge:     rate.SingleMinCharge,
		SingleMaxCharge:     rate.SingleMaxCharge,
		CurrencyCode:        rate.CurrencyCode,
		ApiUrl:              rate.ApiUrl,
		IsRate:              rate.IsRate,
		MerchantPtBalanceId: rate.MerchantPtBalanceId,
	}

	return ProxyPayOrderRequest, rateRpc
}

func (l *ProxyPayOrderLogic) smartProxyPayProcess(merReq *types.ProxyPayRequestX, ProxyPayOrderRequest *transaction.ProxyPayOrderRequest, respOrder *types.OrderX, balanceType *string) (*vo.ProxyPayRespVO, error) {

	orderAmount, _ := merReq.OrderAmount.Float64()
	var MerRateErr error
	var rateList []types.CorrespondMerChnRate
	rateList, MerRateErr = merchantchannelrateservice.SmartGetDesignationMerChnRate(l.svcCtx.MyDB, merReq.MerchantId, constants.CHN_PAY_TYPE_PROXY_PAY, merReq.Currency, orderAmount)

	if len(rateList) > 1 {
		var res *transactionclient.ProxyOrderResponse
		var errBalance error
		for i, _ := range rateList {
			if i == 0 {
				continue
			}
			//1. 将商户钱包加回 (merchantCode, orderNO)
			var resRpc *transaction.ProxyPayFailResponse
			var errRpc error
			if *balanceType == "DFB" {
				resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTransactionFail_DFB(l.ctx, &transaction.ProxyPayFailRequest{
					MerchantCode: respOrder.MerchantCode,
					OrderNo:      respOrder.OrderNo,
				})
			} else if *balanceType == "XFB" {
				resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTransactionFail_XFB(l.ctx, &transaction.ProxyPayFailRequest{
					MerchantCode: respOrder.MerchantCode,
					OrderNo:      respOrder.OrderNo,
				})
			}

			if errRpc != nil {
				logx.WithContext(l.ctx).Errorf("代付智能提单 %s 还款失败。 Err: %s", respOrder.OrderNo, errRpc.Error())
				respOrder.RepaymentStatus = constants.REPAYMENT_FAIL
				if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
					logx.WithContext(l.ctx).Error("代付更新状态错误: ", errUpdate.Error())
				}
				return &vo.ProxyPayRespVO{Code: response.SYSTEM_ERROR, Message: "智能订单流程中斷"}, errorz.New(response.FAIL, errRpc.Error())
			} else {
				logx.WithContext(l.ctx).Infof("智能代付還款rpc完成，%s 錢包還款完成: %#v", balanceType, resRpc)
				respOrder.RepaymentStatus = constants.REPAYMENT_SUCCESS
				if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
					logx.WithContext(l.ctx).Error("代付智能订单更新状态错误: ", errUpdate.Error())
					return &vo.ProxyPayRespVO{Code: response.DATABASE_FAILURE, Message: "代付智能订单更新状态错误"}, errorz.New(response.FAIL, errUpdate.Error())
				}
			}
			// 2. 取得商户对应的代付渠道资料及费率(先收) 计算手续费
			logx.Infof("渠道资讯及费率的计算类型:%+v", rateList[i])
			if MerRateErr != nil {
				logx.WithContext(l.ctx).Error("商户费率错误。%s:%s", MerRateErr.Error(), i18n.Sprintf(MerRateErr.Error()))
				logx.WithContext(l.ctx).Error("商户模式(提單類型 1=多指): ", "提单交易类型:", constants.CHN_PAY_TYPE_PROXY_PAY)
			}

			// 3. 补当渠道成本增加时，如果尚未重新配置商户费率，商户费率小于成本时，会退回提单。
			channelPayType := &types.ChannelPayType{}
			l.svcCtx.MyDB.Table("ch_channel_pay_types").Where("code = ?", rateList[i].ChannelPayTypesCode).Take(channelPayType)
			if rateList[i].Fee != 0 {
				if rateList[i].Fee < channelPayType.Fee {
					logx.WithContext(l.ctx).Errorf("代付提单：%s，商户:%s，代付费率:%f 不可小于渠道成本费率%f ", merReq.OrderNo, merReq.MerchantId, rateList[i].Fee, channelPayType.Fee)
				}
			}

			//4. 取得錢包類型，並執行扣款(因每次扣款錢包類型可能會不同)
			*balanceType, errBalance = merchantbalanceservice.GetBalanceType(l.svcCtx.MyDB, rateList[i].ChannelCode, constants.ORDER_TYPE_DF)
			if errBalance != nil {
				return &vo.ProxyPayRespVO{Code: response.DEDUCTION_FAILED, Message: "查询余额类型错误"}, errBalance
			}
			copierRate := &transaction.CorrespondMerChnRate{}
			copier.Copy(copierRate, rateList[i])

			if *balanceType == "DFB" {
				res, errRpc = l.svcCtx.TransactionRpc.ProxyOrderSmartTranaction_DFB(l.ctx, &transaction.ProxyOrderSmartRequest{
					OrderNo: respOrder.OrderNo,
					Req:     ProxyPayOrderRequest,
					Rate:    copierRate, //智能選取渠道費率
				})
			} else if *balanceType == "XFB" {
				res, errRpc = l.svcCtx.TransactionRpc.ProxyOrderSmartTranaction_XFB(l.ctx, &transaction.ProxyOrderSmartRequest{
					OrderNo: respOrder.OrderNo,
					Req:     ProxyPayOrderRequest,
					Rate:    copierRate, //智能選取渠道費率
				})
			}

			if errRpc != nil {
				logx.WithContext(l.ctx).Error("代付智能提單Rpc呼叫失败:", errRpc.Error())
				//resp.RespCode = response.FAIL
				return &vo.ProxyPayRespVO{Code: response.DEDUCTION_FAILED, Message: "代付智能提單Rpc呼叫失败"}, nil
			} else if res.Code == response.UPDATE_DATABASE_FAILURE {
				return &vo.ProxyPayRespVO{Code: response.DEDUCTION_FAILED, Message: fmt.Sprintf("商户%s余额不足", *balanceType)}, nil
			} else if res.Code == response.DATABASE_FAILURE {
				return &vo.ProxyPayRespVO{Code: response.DEDUCTION_FAILED, Message: fmt.Sprintf("查詢訂單錯誤，單號:%s", respOrder.OrderNo)}, nil
			} else if res.Code != "000" {
				logx.WithContext(l.ctx).Infof("代付智能交易rpc返回失败，%s 錢包扣款失败: %#v", balanceType, res)
				return &vo.ProxyPayRespVO{Code: response.SYSTEM_ERROR, Message: "代付智能交易rpc返回失败"}, errorz.New(res.Code, res.Message)
			} else {
				logx.WithContext(l.ctx).Infof("代付智能交易rpc完成，%s 錢包扣款完成: %#v", balanceType, res)
			}

			var proxyPayRespVO *vo.ProxyPayRespVO
			var errCHN error
			//4. call channel
			proxyPayRespVO, errCHN = ordersService.CallChannel_ProxyOrder(&l.ctx, &l.svcCtx.Config, merReq, respOrder, &rateList[i])
			if errCHN != nil {
				logx.WithContext(l.ctx).Errorf("渠道返回: %s", errCHN)
			}
			logx.WithContext(l.ctx).Infof("渠道返回: %+v", proxyPayRespVO)

			//若"其他錯誤errCHN OR proxyPayRespVO.Code != 202" 則將"渠道返回物件"回傳上層 ，若為202(餘額不足錯誤)則再跑下一組商戶渠道費率
			if proxyPayRespVO != nil && proxyPayRespVO.Code != "202" {
				return proxyPayRespVO, errCHN
			}

		}
	}
	return &vo.ProxyPayRespVO{Code: "202", Message: "渠道余额不足"}, nil
}
