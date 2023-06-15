package merchantbalanceservice

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/types"
	"gorm.io/gorm"
)

// GetBalanceType
// orderType (代付 DF 支付 ZF 下發 XF 內充 NC)
func GetBalanceType(db *gorm.DB, channelCode, orderType string) (balanceType string, err error) {
	balanceType = "XFB"

	// 支付&下發 一定是異動下發餘額
	if orderType == "ZF" || orderType == "XF" {
		return
	}

	// 取得渠道資訊
	var channel types.ChannelData
	if err = db.Table("ch_channels").Where("code = ?", channelCode).Take(&channel).Error; err != nil {
		return "", errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 純代付渠道 代付 內充 異動代付餘額
	if channel.IsProxy == "0" && (orderType == "DF" || orderType == "NC") {
		balanceType = "DFB"
	}

	return
}

/**
  透过订单号取的BalanceType 钱包余额类型
  除了下单之外，还款都要透过订单取得
*/
func GetBalanceTypeByOrder(db *gorm.DB, orderNo string) (balanceType string, err error) {
	balanceType = "XFB"

	// 取得渠道資訊
	var order types.OrderX
	if err = db.Table("tx_orders").Where("order_no = ?", orderNo).Take(&order).Error; err != nil {
		return "", errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return order.BalanceType, nil
}
