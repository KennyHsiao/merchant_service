package transactionLogService

import (
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/merchant/internal/types"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"time"
)

//交易日志新增Func
func CreateTransactionLog(db *gorm.DB, data *types.TransactionLogData) (err error) {

	jsonContent, err := json.Marshal(data.Content)
	if err != nil {
		logx.Errorf("產生交易日志錯誤:%s", err.Error())
	}

	txLog := types.TxLog{
		MerchantCode:    data.MerchantNo,
		MerchantOrderNo: data.MerchantOrderNo,
		OrderNo:         data.OrderNo,
		LogType:         data.LogType,
		LogSource:       data.LogSource,
		Content:         string(jsonContent),
		Log:             ProduceLogFromTemplate(data, string(jsonContent)), //日誌說明(Template)
		CreatedAt:       time.Now().UTC().String(),
		TraceId:         data.TraceId,
	}

	go func() {
		if err = db.Table("tx_log").Create(&txLog).Error; err != nil {
			return
		}
	}()

	return nil
}

func ProduceLogFromTemplate(data *types.TransactionLogData, jsonStr string) (log string) {
	//LogType
	//ERROR_MSG                 = "1" //1:錯誤訊息
	//MERCHANT_REQUEST          = "2" //2:商户请求
	//ERROR_REPLIED_TO_MERCHANT = "3" //3:返回商户错误
	//DATA_REQUEST_CHANNEL      = "4" //4.打给渠道资料
	//RESPONSE_FROM_CHANNEL     = "5" //5.渠道返回资料
	//CALLBACK_FROM_CHANNEL     = "6" //6.渠道回调资料
	//CALLBACK_TO_MERCHANT      = "7" //7.回调给商户

	var err error
	if data.LogSource == constants.API_DF {
		switch data.LogType {
		case constants.MERCHANT_REQUEST:
			var req types.ProxyPayRequestX
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)
			log = fmt.Sprintf(constants.PATTERN_1, req.NotifyUrl, req.BankName, req.BankNo, req.DefrayName, req.Currency, req.OrderAmount)

		case constants.CALLBACK_FROM_CHANNEL:
			var req types.ProxyPayOrderCallBackRequest
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)
			log = fmt.Sprintf(constants.PATTERN_5, req.ChannelResultStatus, req.ChannelResultNote, req.Amount, req.ChannelCharge)

		case constants.CALLBACK_TO_MERCHANT:
			var req interface{}
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)
			myMap := req.(map[string]interface{})
			log = fmt.Sprintf(constants.PATTERN_7, myMap["orderAmount"], myMap["orderStatus"])

		case constants.ERROR_REPLIED_TO_MERCHANT:
			contentStrut := struct {
				ErrorCode string
				ErrorMsg  string
			}{}
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &contentStrut)
			log = fmt.Sprintf(constants.MERCHANT_ERROR_RESPONSE, contentStrut.ErrorCode, contentStrut.ErrorMsg)
		}

	} else if data.LogSource == constants.API_ZF {
		switch data.LogType {
		case constants.MERCHANT_REQUEST:
			var req types.PayOrderRequestX
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)
			log = fmt.Sprintf(constants.PATTERN_2, req.NotifyUrl, req.Currency, req.PayType, req.PayTypeNo, req.OrderAmount)

		case constants.CALLBACK_FROM_CHANNEL:
			var req types.PayCallBackRequest
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)
			log = fmt.Sprintf(constants.PATTERN_4, req.OrderAmount, req.OrderStatus)

		case constants.CALLBACK_TO_MERCHANT:
			var req vo.PayCallBackVO
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)
			log = fmt.Sprintf(constants.PATTERN_6, req.OrderAmount, req.OrderStatus)

		case constants.ERROR_REPLIED_TO_MERCHANT:
			var contentStrut struct {
				ErrorCode string
				ErrorMsg  string
				//Resp      interface{}
			}
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &contentStrut)
			log = fmt.Sprintf(constants.MERCHANT_ERROR_RESPONSE, contentStrut.ErrorCode, contentStrut.ErrorMsg)
		}
	} else if data.LogSource == constants.API_XF {
		switch data.LogType {
		case constants.MERCHANT_REQUEST:
			var req types.WithdrawApiOrderRequest
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", data.Content)), &req)
			log = fmt.Sprintf(constants.PATTERN_3, req.NotifyUrl, req.WithdrawName, req.BankName, req.AccountNo, req.OrderAmount)
			//收款人：%s,银行名称：%s,卡号：%s,下发金额：%s"
		case constants.CALLBACK_TO_MERCHANT:
		}
	}
	if err != nil {
		logx.Errorf("產生交易日誌模板錯誤:", err.Error())
	}

	return
}

//交易日志新增Func
//func CreateTransactionLog(db *gorm.DB, orderX *types.OrderX, data *types.TransactionLogData) (err error) {
//
//	var logSource string
//	if orderX.Source == constants.UI {
//		if orderX.Type == constants.ORDER_TYPE_DF {
//			logSource = constants.PLATEFORM_DF
//		} else if orderX.Type == constants.ORDER_TYPE_NC {
//			logSource = constants.PLATEFORM_NC
//		} else if orderX.Type == constants.ORDER_TYPE_XF {
//			logSource = constants.PLATEFORM_XF
//		}
//	} else if orderX.Source == constants.API {
//		if orderX.Type == constants.ORDER_TYPE_DF {
//			logSource = constants.API_DF
//		} else if orderX.Type == constants.ORDER_TYPE_ZF {
//			logSource = constants.API_ZF
//		} else if orderX.Type == constants.ORDER_TYPE_XF {
//			logSource = constants.API_XF
//		}
//	}
//
//	txLog := types.TxLog{
//		MerchantCode:    orderX.MerchantCode,
//		OrderNo:         orderX.OrderNo,
//		MerchantOrderNo: orderX.MerchantOrderNo,
//		ChannelOrderNo:  orderX.ChannelOrderNo,
//		LogType:         data.LogType,
//		LogSource:       logSource,
//		Content:         data.Content,
//		CreatedAt:       time.Now().UTC().String(),
//		ErrorCode:       data.ErrCode,
//		ErrorMsg:        data.ErrMsg,
//	}
//
//	if err = db.Table("tx_log").Create(&txLog).Error; err != nil {
//		return
//	}
//
//	return nil
//}
