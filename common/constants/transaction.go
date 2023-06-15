package constants

const (
	//交易日志类型
	ERROR_MSG                 = "1" //1:錯誤訊息
	MERCHANT_REQUEST          = "2" //2:商户请求
	ERROR_REPLIED_TO_MERCHANT = "3" //3:返回商户错误
	DATA_REQUEST_CHANNEL      = "4" //4.打给渠道资料
	RESPONSE_FROM_CHANNEL     = "5" //5.渠道返回资料
	CALLBACK_FROM_CHANNEL     = "6" //6.渠道回调资料
	CALLBACK_TO_MERCHANT      = "7" //7.回调给商户
	RESPONSE_FROM_MERCHANT    = "8" //8.商戶返回訊息

	//日誌來源(1:內充平台、2:支付API、3:代付API、4:代付平台、5:下發API)
	PLATEFORM_NC = "1"
	API_ZF       = "2"
	API_DF       = "3"
	PLATEFORM_DF = "4"
	API_XF       = "5"
	PLATEFORM_XF = "6"

	PATTERN_1 = "通知路径 : %s</br>开户行行名：%s,银行卡号：%s,开户人姓名：%s</br>币别 : %s,代付金额：%s" //商户请求	代付API
	PATTERN_2 = "通知路径 : %s</br>币别 : %s,支付类型 : %s,支付渠道指定代码 : %s</br>订单金额：%s"   //商户请求	支付API
	PATTERN_3 = "通知路径 : %s</br>收款人：%s,银行名称：%s,卡号：%s,下发金额：%s"                  //商户请求	下发API

	PATTERN_4 = "实付金额：%f,支付结果：%s"                     //渠道回调数据	支付API
	PATTERN_5 = "渠道结果状态：%s,渠道结果说明：%s,代付金额：%f,渠道费用：%f" //渠道回调数据	代付API

	PATTERN_6               = "订单金额：%s,订单状态：%s"          //回调给商户	支付API
	PATTERN_7               = "订单金额：%s,订单状态：%s"          //回调给商户	代付API
	PATTERN_8               = "订单金额：%s,手續費：%s ,订单状态：%s"  //回调给商户	下发API
	MERCHANT_ERROR_RESPONSE = "ErrorCode:%s ErrorMsg:%s" //返回商户错误

)
