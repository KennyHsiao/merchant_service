package bo

type PayBO struct {
	MerchantOrderNo   string `json:"merchantOrderNo"`
	OrderNo           string `json:"orderNo"`
	PayType           string `json:"payType"`
	ChannelPayType    string `json:"channelPayType"`
	TransactionAmount string `json:"transactionAmount"`
	BankCode          string `json:"bankCode"`
	PageUrl           string `json:"pageUrl"`
	OrderName         string `json:"orderName"`
	MerchantId        string `json:"merchantId"`
	Currency          string `json:"currency"`
	SourceIp          string `json:"sourceIp"`
	UserId            string `json:"userId"`
	JumpType          string `json:"jumpType"`
	PlayerId          string `json:"playerId"`
	Address           string `json:"address, optional"`
	City              string `json:"city, optional"`
	ZipCode           string `json:"zipCode, optional"`
	Country           string `json:"country, optional"`
	Phone             string `json:"phone, optional"`
	Email             string `json:"email, optional"`
	PageFailedUrl     string `json:"pageFailedUrl, optional"`
}
