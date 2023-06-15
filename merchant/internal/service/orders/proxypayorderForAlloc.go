package ordersService

import (
	"fmt"
	"github.com/neccoys/go-zero-extension/redislock"
	"time"

	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/types"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/zeromicro/go-zero/core/logx"
)

func (l *OrdersService) ChannelCallBackAlloc(req *types.ProxyPayOrderCallBackRequest, orderX *types.OrderX) error {
	redisKey := fmt.Sprintf("%s-%s", orderX.MerchantCode, req.ProxyPayOrderNo)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "proxy-call-back:")
	redisLock.SetExpire(5)
	if isOK, _ := redisLock.Acquire(); isOK {
		if err := l.internalChannelCallBackForAlloc(req, orderX); err != nil {
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
	渠道(代付撥款)回调用，更新状态用(Service)
	提单order_status为[1=成功][2=失败]时，不接受回调的变更
*/
func (l *OrdersService) internalChannelCallBackForAlloc(req *types.ProxyPayOrderCallBackRequest, orderX *types.OrderX) error {

	//订单状态为[成功]或[失败]，判定为已结单，以及[人工處理單]不接受回調
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
		orderX.RepaymentStatus = constants.REPAYMENT_NOT //还款状态：(0：不需还款、[1]：待还款、2：还款成功、3：还款失败)，预设不需还款
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

	// 更新订单
	if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(orderX).Error; errUpdate != nil {
		logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
	}

	return nil
}
