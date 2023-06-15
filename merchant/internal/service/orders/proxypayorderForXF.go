package ordersService

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/service/merchantbalanceservice"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	"com.copo/bo_service/merchant/internal/types"
	"errors"
	"fmt"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/neccoys/go-zero-extension/redislock"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"time"
)

func (l *OrdersService) ChannelCallBackForXF(req *types.ProxyPayOrderCallBackRequest, orderX *types.OrderX) error {
	redisKey := fmt.Sprintf("%s-%s", orderX.MerchantCode, req.ProxyPayOrderNo)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "proxy-XF-call-back:")
	redisLock.SetExpire(5)
	if isOK, _ := redisLock.Acquire(); isOK {
		if err := l.internalChannelCallBackForXF(req, orderX); err != nil {
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
func (l *OrdersService) internalChannelCallBackForXF(req *types.ProxyPayOrderCallBackRequest, orderX *types.OrderX) error {

	OrderChannelX := types.OrderChannelsX{}

	if err := l.svcCtx.MyDB.Table("tx_order_channels").
		Where("order_sub_no = ?", req.ProxyPayOrderNo).
		Take(OrderChannelX).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorz.New(response.DATA_NOT_FOUND)
		} else {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	//订单状态为[成功]或[失败]，判定为已结单，以及[人工處理單]不接受回調
	if OrderChannelX.Status == constants.SUCCESS || OrderChannelX.Status == constants.FAIL || OrderChannelX.Status == constants.FROZEN ||
		orderX.Status == constants.SUCCESS || orderX.Status == constants.FAIL || orderX.Status == constants.FROZEN {
		logx.WithContext(l.ctx).Infof("代付订单：%s，提单状态为：%s，判定为已结单，不接受回调变更。", OrderChannelX.OrderNo, OrderChannelX.Status)
		//TODO 是否有需要回寫代付歷程?
		return errorz.New(response.PROXY_PAY_IS_CLOSE, "此提单目前已为结单状态")
	} else if orderX.PersonProcessStatus != constants.PERSON_PROCESS_STATUS_NO_ROCESSING {
		logx.WithContext(l.ctx).Infof("代付订单：%s，人工处里状态为：%s，判定已进入人工处里阶段，不接受回调变更。", orderX.OrderNo, orderX.PersonProcessStatus)
		//TODO 是否有需要回寫代付歷程?
		return errorz.New(response.PROXY_PAY_IS_PERSON_PROCESS, "提单目前为人工处里阶段，不可回调变更")
	} else if req.ChannelResultStatus == constants.CALL_BACK_STATUS_SUCCESS && req.Amount != OrderChannelX.OrderAmount {
		//代付回调金额若不等于订单金额，订单转 人工处理，并塞error_message
		logx.WithContext(l.ctx).Errorf("代付回调金额不等于订单金额，订单转人工处理。单号:%s", orderX.OrderNo)
		OrderChannelX.ErrorMsg = "代付渠道回调金额不等于订单金额"
		//orderX.ErrorNote = "代付渠道回调金额不等于订单金额"
		//orderX.ErrorType = "1"
		//orderX.PersonProcessStatus = constants.PERSON_PROCESS_WAIT_CHANGE
		//orderX.RepaymentStatus = constants.REPAYMENT_WAIT

		if errUpdate := l.svcCtx.MyDB.Table("tx_order_channels").Updates(OrderChannelX).Error; errUpdate != nil {
			logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
		}

		return errorz.New(response.PROXY_PAY_AMOUNT_INVALID, "代付回调金额与订单金额不符")
	}
	/*更新訂單:
	1. 訂單狀態(依渠道回調決定)
	2. 還款狀態(預設[0]：不需还款)，若渠道回調失敗單，則[1]：待还款
	*/
	//订单预设还款状态为"不需还款"，更新為待还款
	OrderChannelX.Status = getOrderStatus(req.ChannelResultStatus)
	//orderX.RepaymentStatus = constants.REPAYMENT_NOT //还款状态：([0]：不需还款、1：待还款、2：还款成功、3：还款失败)，预设不需还款
	if req.ChannelResultStatus == constants.CALL_BACK_STATUS_FAIL {
		//orderX.RepaymentStatus = constants.REPAYMENT_WAIT //还款状态：(0：不需还款、[1]：待还款、2：还款成功、3：还款失败)，预设不需还款
		if req.ChannelResultNote != "" {
			OrderChannelX.ErrorMsg = "渠道回调:-" + req.ChannelResultNote //失败时，写入渠道返还的讯息
		} else {
			OrderChannelX.ErrorMsg = "渠道回调: 交易失败"
		}
	}

	OrderChannelX.TransAt = types.JsonTime{}.New()
	OrderChannelX.UpdatedAt = time.Now().UTC()

	if OrderChannelX.ChannelOrderNo == "" {
		OrderChannelX.ChannelOrderNo = req.ChannelOrderNo
	}

	if req.ChannelCharge != 0 {
		//渠道有回傳渠道手續費
	}

	// 更新订单
	if errUpdate := l.svcCtx.MyDB.Table("tx_order_channels").Updates(OrderChannelX).Error; errUpdate != nil {
		logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
	}

	// 更新訂單訂單歷程 (不抱錯)
	if OrderChannelX.Status == constants.SUCCESS {
		if err4 := l.svcCtx.MyDB.Table("tx_order_actions").Create(&types.OrderActionX{
			OrderAction: types.OrderAction{
				OrderNo:     OrderChannelX.OrderNo,
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
	if orderX.IsMerchantCallback == constants.IS_MERCHANT_CALLBACK_NO {
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
