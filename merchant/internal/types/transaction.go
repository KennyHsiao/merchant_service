package types

import (
	"com.copo/bo_service/common/gormx"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"
	"time"
	"unsafe"
)

type JsonTime time.Time

func (Merchant) TableName() string {
	return "mc_merchants"
}

func (MerchantCurrency) TableName() string {
	return "mc_merchant_currencies"
}

func (MerchantBalance) TableName() string {
	return "mc_merchant_balances"
}

func (o MerchantContact) Value() (driver.Value, error) {
	b, err := json.Marshal(o)
	return string(b), err
}

func (o *MerchantContact) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), o)
}

func (o MerchantBizInfo) Value() (driver.Value, error) {
	b, err := json.Marshal(o)
	return string(b), err
}

func (o *MerchantBizInfo) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), o)
}

func (j JsonTime) MarshalJSON() ([]byte, error) {
	var res string
	if !time.Time(j).IsZero() {
		var stamp = fmt.Sprintf("%s", time.Time(j).Format("2006-01-02 15:04:05"))
		str := strings.Split(stamp, " +")
		res = str[0]
		return json.Marshal(res)
	}
	return json.Marshal("")
}

func (j JsonTime) Time() time.Time {
	return time.Time(j)
}

func (j JsonTime) Value() (driver.Value, error) {
	return time.Time(j), nil
}

func (j JsonTime) Parse(s string, zone ...string) (JsonTime, error) {

	var (
		loc *time.Location
		err error
	)
	if len(zone) > 0 {
		loc, err = time.LoadLocation(zone[0])
	} else {
		loc, err = time.LoadLocation("")
	}

	if err != nil {
		return j, err
	}

	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, loc)
	if err != nil {
		return j, err
	}
	jt := (*JsonTime)(unsafe.Pointer(&t))

	return *jt, nil
}

func (j JsonTime) New(ts ...time.Time) JsonTime {
	var t time.Time

	if len(ts) > 0 {
		t = ts[0]
	} else {
		t = time.Now().UTC()
	}

	jt := (*JsonTime)(unsafe.Pointer(&t))
	return *jt
}

func (OrderChannels) TableName() string {
	return "tx_order_channels"
}

type OrderX struct {
	Order
	TransAt            JsonTime  `json:"transAt, optional"`
	MerchantCallBackAt time.Time `json:"merchantCallBackAt"`
	ChannelCallBackAt  time.Time `json:"channelCallBackAt"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type OrderInternalCreate struct {
	OrderX
	FormData map[string][]*multipart.FileHeader `gorm:"-"`
}

type OrderInternalUpdate struct {
	OrderX
}

type OrderWithdrawUpdate struct {
	OrderX
}

type MerchantRateListViewX struct {
	MerchantRateListView
	Balance float64 `json:"balance"`
}

type MerchantOrderRateListViewX struct {
	MerchantOrderRateListView
	Balance float64 `json:"balance"`
}

type OrderQueryMerchantCurrencyAndBanksResponseX struct {
	MerchantOrderRateListViewX *MerchantOrderRateListViewX `json:"merchantOrderRateListViewX"`
	ChannelBanks               []ChannelBankX
}

type ReceiptRecordQueryAllRequestX struct {
	ReceiptRecordQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type FrozenRecordQueryAllRequestX struct {
	FrozenRecordQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type DeductRecordQueryAllRequestX struct {
	DeductRecordQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type WithdrawOrderUpdateRequestX struct {
	List []ChannelWithdraw `json:"list"`
	OrderX
}

type OrderActionX struct {
	OrderAction
	CreatedAt time.Time
}

type OrderChannelsX struct {
	OrderChannels
	CreatedAt time.Time
	UpdatedAt time.Time
	TransAt   JsonTime `json:"transAt, optional"`
}
type OrderFeeProfitX struct {
	OrderFeeProfit
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ReceiptRecordX struct {
	ReceiptRecord
	TransAt JsonTime `json:"transAt, optional"`
}

type FrozenRecordX struct {
	FrozenRecord
	TransAt JsonTime `json:"transAt, optional"`
}

type DeductRecordX struct {
	DeductRecord
	TransAt JsonTime `json:"trans_at, optional"`
}

type ReceiptRecordQueryAllResponseX struct {
	List     []ReceiptRecordX `json:"list"`
	PageNum  int              `json:"pageNum" gorm:"-"`
	PageSize int              `json:"pageSize" gorm:"-"`
	RowCount int64            `json:"rowCount"`
}

type FrozenRecordQueryAllResponseX struct {
	List     []FrozenRecordX `json:"list"`
	PageNum  int             `json:"pageNum" gorm:"-"`
	PageSize int             `json:"pageSize" gorm:"-"`
	RowCount int64           `json:"rowCount"`
}

type DeductRecordQueryAllResponseX struct {
	List     []DeductRecordX `json:"list"`
	PageNum  int             `json:"pageNum" gorm:"-"`
	PageSize int             `json:"pageSize" gorm:"-"`
	RowCount int64           `json:"rowCount"`
}

type OrderActionQueryAllRequestX struct {
	OrderActionQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type PersonalRepaymentRequestX struct {
	PersonalRepaymentRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type PersonalRepaymentResponseX struct {
	List     []PersonalRepayment `json:"list"`
	PageNum  int                 `json:"pageNum" gorm:"-"`
	PageSize int                 `json:"pageSize" gorm:"-"`
	RowCount int64               `json:"rowCount"`
}

type PersonalRepaymentX struct {
	PersonalRepayment
	TransAt JsonTime `json:"transAt, optional"`
}

type PersonalStatusUpdateResponseX struct {
	PersonalStatusUpdateResponse
	ChannelTransAt JsonTime `json:"channelTransAt, optional"`
}

type OrderFeeProfitQueryAllRequestX struct {
	OrderFeeProfitQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type CalculateProfit struct {
	MerchantCode        string
	OrderNo             string
	Type                string
	CurrencyCode        string
	BalanceType         string
	ChannelCode         string
	ChannelPayTypesCode string
	OrderAmount         float64
}

type PayOrderRequestX struct {
	PayOrderRequest
	AccessType  json.Number `json:"accessType, optional" validate:"required"`
	PayTypeNo   json.Number `json:"payTypeNo, optional"`
	OrderAmount json.Number `json:"orderAmount, optional" validate:"required,jsanNumPrec=2"`
	MyIp        string      `json:"my_ip, optional"`
}

type PayQueryRequestX struct {
	PayQueryRequest
	AccessType json.Number `json:"accessType"`
	MyIp       string      `json:"my_ip, optional"`
}

type PayQueryBalanceRequestX struct {
	PayQueryBalanceRequest
	AccessType json.Number `json:"accessType"`
	MyIp       string      `json:"my_ip, optional"`
}

type PayQuerySubBalanceRequestX struct {
	PayQuerySubBalanceResponse
	AccessType json.Number `json:"accessType"`
	MyIp       string      `json:"myIp"`
}
type ProxyPayRequestX struct {
	ProxyPayOrderRequest
	AccessType  json.Number `json:"accessType"`
	OrderAmount json.Number `json:"orderAmount, optional" validate:"required"` //到小數兩位
	Ip          string      `json:"ip, optional"`
}

type ProxyPayOrderQueryRequestX struct {
	ProxyPayOrderQueryRequest
	AccessType json.Number `json:"accessType"`
	Ip         string      `json:"ip, optional"`
}

type WithdrawApiOrderRequestX struct {
	WithdrawApiOrderRequest
	OrderAmount json.Number `json:"orderAmount, optional"`
	//PtBalanceId json.Number `json:"ptBalanceId, optional"`
	MyIp string `json:"my_ip, optional"`
}

type MultipleOrderWithdrawCreateRequestX struct {
	List []OrderWithdrawCreateRequestX `json:"list"`
}

type OrderWithdrawCreateRequestX struct {
	OrderWithdrawCreateRequest
	MerchantCode    string `json:"merchantCode, optional"`
	MerchantOrderNo string `json:"merchant_order_no, optional"`
	UserAccount     string `json:"userAccount, optional"`
	NotifyUrl       string `json:"notify_url, optional"`
	PageUrl         string `json:"page_url, optional"`
	Source          string `json:"source, optional"`
	Type            string `json:"type, optional"`
	ChangeType      string `json:"change_type, optional"`
	PtBalanceId     int64  `json:"ptBalanceId, optional"`
}

type OrderWithdrawCreateResponse struct {
	OrderX
	OrderNo string   `json:"orderNo"`
	Index   []string `json:"index"`
	Errs    []string `json:"errs"`
}

type WithdrawApiQueryRequestX struct {
	WithdrawApiQueryRequest
	MyIp string `json:"my_ip, optional"`
}

//以下補struct

type ChannelData struct {
	ID                      int64            `json:"id, optional"`
	Code                    string           `json:"code, optional"`
	Name                    string           `json:"name, optional"`
	ProjectName             string           `json:"projectName, optional"`
	IsProxy                 string           `json:"isProxy, optional"`
	IsNzPre                 string           `json:"isNzPre, optional"`
	ApiUrl                  string           `json:"apiUrl, optional"`
	CurrencyCode            string           `json:"currencyCode, optional"`
	ChannelWithdrawCharge   float64          `json:"channelWithdrawCharge, optional"`
	Balance                 float64          `json:"balance, optional"`
	Status                  string           `json:"status, optional"`
	Device                  string           `json:"device,optional"`
	MerId                   string           `json:"merId, optional"`
	MerKey                  string           `json:"merKey, optional"`
	PayUrl                  string           `json:"payUrl, optional"`
	PayQueryUrl             string           `json:"payQueryUrl, optional"`
	PayQueryBalanceUrl      string           `json:"payQueryBalanceUrl, optional"`
	ProxyPayUrl             string           `json:"proxyPayUrl, optional"`
	ProxyPayQueryUrl        string           `json:"proxyPayQueryUrl, optional"`
	ProxyPayQueryBalanceUrl string           `json:"proxyPayQueryBalanceUrl, optional"`
	WhiteList               string           `json:"whiteList, optional"`
	PayTypeMapList          []PayTypeMap     `json:"payTypeMapList, optional" gorm:"-"`
	PayTypeMap              string           `json:"payTypeMap, optional"`
	ChannelPayTypeList      []ChannelPayType `json:"channelPayTypeList, optional" gorm:"foreignKey:ChannelCode;references:Code"`
	ChannelPort             string           `json:"channelPort, optional"`
	WithdrawBalance         float64          `json:"withdrawBalance, optional"`
	ProxypayBalance         float64          `json:"proxypayBalance, optional"`
	BankCodeMapList         []BankCodeMap    `json:"bankCodeMapList, optional" gorm:"-"`
	Banks                   []Bank           `json:"banks, optional" gorm:"many2many:ch_channel_banks;foreignKey:Code;joinForeignKey:channel_code;references:bank_no;joinReferences:bank_no"`
}

type PayTypeMap struct {
	ID      int64  `json:"id, optional"`
	PayType string `json:"payType"`
	TypeNo  string `json:"typeNo"`
	MapCode string `json:"mapCode"`
}

type ChannelPayType struct {
	ID                int64   `json:"id, optional"`
	Code              string  `json:"code, optional"`
	ChannelCode       string  `json:"channelCode, optional"`
	PayTypeCode       string  `json:"payTypeCode, optional"`
	Fee               float64 `json:"fee, optional"`
	HandlingFee       float64 `json:"handlingFee, optional"`
	MaxInternalCharge float64 `json:"maxInternalCharge, optional"`
	DailyTxLimit      float64 `json:"dailyTxLimit, optional"`
	SingleMinCharge   float64 `json:"singleMinCharge, optional"`
	SingleMaxCharge   float64 `json:"singleMaxCharge, optional"`
	FixedAmount       string  `json:"fixedAmount, optional"`
	BillDate          int64   `json:"billDate, optional"`
	Status            string  `json:"status, optional"`
	IsProxy           string  `json:"isProxy, optional"`
	Device            string  `json:"device, optional"`
	MapCode           string  `json:"mapCode, optional"`
}

type BankCodeMap struct {
	ID          int64  `json:"id, optional"`
	ChannelCode string `json:"channelCode, optional"`
	BankNo      string `json:"bankNo"`
	MapCode     string `json:"mapCode"`
}

type ChannelBankX struct {
	ChannelCode string `json:"channel_code"`
	BankNo      string `json:"bank_no"`
	BankName    string `json:"bank_name"`
}

type FrozenRecordQueryAllRequest struct {
	MerchantCode    string `json:"merchantCode, optional"`
	OrderNo         string `json:"orderNo, optional"`
	MerchantOrderNo string `json:"merchantOrderNo, optional"`
	StartAt         string `json:"startAt, optional"`
	EndAt           string `json:"endAt, optional"`
	Type            string `json:"type, optional"`
	PayTypeCode     string `json:"payTypeCode, optional"`
	ChannelCode     string `json:"channelCode, optional"`
	CurrencyCode    string `json:"currencyCode, optional"`
	Source          string `json:"source, optional"`
	PageNum         int    `json:"pageNum, optional" gorm:"-"`
	PageSize        int    `json:"pageSize, optional" gorm:"-"`
}

type OrderFeeProfit struct {
	ID                  int64   `json:"id"`
	OrderNo             string  `json:"orderNo"`
	MerchantCode        string  `json:"merchantCode"`
	AgentLayerNo        string  `json:"agentLayerNo"`
	AgentParentCode     string  `json:"agentParentCode"`
	BalanceType         string  `json:"balanceType"`
	Fee                 float64 `json:"fee"`
	HandlingFee         float64 `json:"handlingFee"`
	TransferHandlingFee float64 `json:"transferHandlingFee"`
	ProfitAmount        float64 `json:"profitAmount"`
}

type PayType struct {
	ID           int64         `json:"id, optional"`
	Code         string        `json:"code, optional"`
	Name         string        `json:"name, optional"`
	Currency     string        `json:"currency, optional"`
	ImgUrl       string        `json:"imgUrl, optional"`
	ChannelDatas []ChannelData `json:"channelDatas, optional" gorm:"many2many:ch_channel_pay_types:foreignKey:code;joinForeignKey:pay_type_code;references:code;joinReferences:channel_code"`
}

type PersonalRepaymentRequest struct {
	OrderNo      string `json:"orderNo, optional"`
	ChannelCode  string `json:"channelCode, optional"`
	Source       string `json:"source"`
	StartAt      string `json:"startAt"`
	EndAt        string `json:"endAt"`
	CurrencyCode string `json:"currencyCode"`
	PageNum      int    `json:"pageNum, optional" gorm:"-"`
	PageSize     int    `json:"pageSize, optional" gorm:"-"`
}

type PersonalStatusUpdateResponse struct {
	OrderNo             string `json:"orderNo"`
	PersonProcessStatus string `json:"personProcessStatus"`
	Comment             string `json:"comment, optional"`
	ChannelOrderNo      string `json:"channelOrderNo, optional"`
	ChannelTransAt      string `json:"channelTransAt, optional"`
}

type OrderFeeProfitQueryAllRequest struct {
	OrderNo         string `json:"orderNo"`
	MerchantCode    string `json:"merchantCode"`
	AgentLayerNo    string `json:"agentLayerNo"`
	AgentParentCode string `json:"agentParentCode"`
	PageNum         int    `json:"pageNum, optional" gorm:"-"`
	PageSize        int    `json:"pageSize, optional" gorm:"-"`
}

type CorrespondMerChnRate struct {
	MerchantCode        string  `json:"merchantCode"`
	ChannelPayTypesCode string  `json:"channelPayTypesCode"`
	ChannelCode         string  `json:"channelCode"`
	PayTypeCode         string  `json:"payTypeCode"`
	Designation         string  `json:"designation"`
	DesignationNo       string  `json:"designationNo"`
	Fee                 float64 `json:"fee"`
	HandlingFee         float64 `json:"handlingFee"`
	MapCode             string  `json:"map_code"`
	ChFee               float64 `json:"chFee"`
	ChHandlingFee       float64 `json:"chHandlingFee"`
	MerchantPtBalanceId int64   `json:"merchantPtBalanceId"`
	SingleMinCharge     float64 `json:"singleMinCharge"`
	SingleMaxCharge     float64 `json:"singleMaxCharge"`
	CurrencyCode        string  `json:"currencyCode"`
	ApiUrl              string  `json:"apiUrl"`
	ChannelPort         string  `json:"channelPort"`
	IsRate              string  `json:"isRate"`
}

type UpdateFrozenAmount struct {
	MerchantCode    string
	CurrencyCode    string
	OrderNo         string
	OrderType       string
	ChannelCode     string
	PayTypeCode     string
	TransactionType string
	BalanceType     string
	FrozenAmount    float64
	Comment         string
	CreatedBy       string
}

type MerchantBalanceX struct {
	MerchantBalance
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UpdateBalance struct {
	MerchantCode    string
	CurrencyCode    string
	OrderNo         string
	OrderType       string
	ChannelCode     string
	PayTypeCode     string
	TransactionType string
	BalanceType     string
	TransferAmount  float64
	Comment         string
	CreatedBy       string
}
type MerchantBalanceRecordX struct {
	MerchantBalanceRecord
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MerchantFrozenRecordX struct {
	MerchantFrozenRecord
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (o MenuMeta) Value() (driver.Value, error) {
	b, err := json.Marshal(o)
	return string(b), err
}

func (o *MenuMeta) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), o)
}

type MenuTree struct {
	Menu
	Child []*MenuTree `json:"child"`
}

type Menu struct {
	ID        int64    `json:"id"`
	ParentID  int64    `json:"parentId, optional"`
	Name      string   `json:"name, optional"`
	Group     string   `json:"group, optional"`
	Component string   `json:"component, optional"`
	Path      string   `json:"path, optional"`
	Meta      MenuMeta `json:"meta, optional"`
	Hidden    bool     `json:"hidden, optional"`
	SortOrder int64    `json:"sortOrder, optional"`
	Permits   []Permit `json:"permits, optional" gorm:"foreignKey:menu_id;"`
}

type MenuMeta struct {
	Icon    string `json:"icon"`
	Title   string `json:"title"`
	NoCache bool   `json:"noCache"`
}

type MenuCreateRequest struct {
	ID        int64    `json:"id, optional"`
	ParentID  int64    `json:"parentId, optional"`
	Name      string   `json:"name" valiate:"required,lte=32"`
	Group     string   `json:"group" valiate:"required,lte=32"`
	Component string   `json:"component" valiate:"required,lte=128"`
	Path      string   `json:"path" valiate:"required,lte=255"`
	Meta      MenuMeta `json:"meta, optional"`
	Hidden    bool     `json:"hidden"`
	SortOrder int64    `json:"sortOrder, optional"`
	Permits   []Permit `json:"permits, optional" gorm:"foreignKey:menu_id;"`
}

type MenuUpdateRequest struct {
	ID        int64    `json:"id"`
	ParentID  int64    `json:"parentId, optional"`
	Name      string   `json:"name" valiate:"required,lte=32"`
	Group     string   `json:"group" valiate:"required,lte=32"`
	Component string   `json:"component" valiate:"required,lte=128"`
	Path      string   `json:"path" valiate:"required,lte=255"`
	Meta      MenuMeta `json:"meta, optional"`
	Hidden    bool     `json:"hidden"`
	SortOrder int64    `json:"sortOrder, optional"`
	Permits   []Permit `json:"permits, optional" gorm:"foreignKey:menu_id;"`
}

type MenuDeleteRequest struct {
	ID int64 `json:"id"`
}

type MenuQueryRequest struct {
	ID int64 `json:"id"`
}

type MenuQueryResponse struct {
	Menu
}

type MenuQueryAllRequest struct {
	Name     string `json:"name, optional"`
	Group    string `json:"group, optional"`
	PageNum  int    `json:"pageNum" gorm:"-"`
	PageSize int    `json:"pageSize" gorm:"-"`
}

type MenuQueryAllResponse struct {
	List     []Menu `json:"list"`
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
	RowCount int64  `json:"rowCount"`
}

type Permit struct {
	ID           int64  `json:"id, optional"`
	MenuID       int64  `json:"menuId, optional"`
	Slug         string `json:"slug, optional"`
	Name         string `json:"name, optional"`
	NeedPassword bool   `json:"needPassword"`
}

type PermitCreateRequest struct {
	ID           int64  `json:"id, optional"`
	MenuID       int64  `json:"menuId" validate:"required"`
	Slug         string `json:"slug" validate:"required,lte=16"`
	Name         string `json:"name" validate:"required,lte=32"`
	NeedPassword bool   `json:"needPassword"`
}

type PermitUpsertRequest struct {
	MenuID  int64    `json:"menuId"`
	Permits []Permit `json:"permits"`
}

type PermitUpdateRequest struct {
	ID           int64  `json:"id"`
	MenuID       int64  `json:"menuId" validate:"required"`
	Slug         string `json:"slug" validate:"required,lte=16"`
	Name         string `json:"name" validate:"required,lte=32"`
	NeedPassword bool   `json:"needPassword"`
}

type PermitDeleteRequest struct {
	ID int64 `json:"id"`
}

type PermitQueryRequest struct {
	ID int64 `json:"id"`
}

type PermitQueryResponse struct {
	Permit
}

type PermitQueryAllRequest struct {
	MenuID   int64 `json:"menuId, optional"`
	PageNum  int   `json:"pageNum" gorm:"-"`
	PageSize int   `json:"pageSize" gorm:"-"`
}

type PermitQueryAllResponse struct {
	List     []Permit `json:"list"`
	PageNum  int      `json:"pageNum"`
	PageSize int      `json:"pageSize"`
	RowCount int64    `json:"rowCount"`
}

type Role struct {
	ID      int64    `json:"id, optional"`
	Slug    string   `json:"slug, optional"`
	Name    string   `json:"name, optional"`
	Menus   []Menu   `json:"menus, optional" gorm:"many2many:au_role_menus;foreignKey:id;joinForeignKey:role_id;"`
	Permits []Permit `json:"permits, optional" gorm:"many2many:au_role_permits;foreignKey:id;joinForeignKey:role_id;"`
}

type RoleCreateRequest struct {
	ID      int64    `json:"id, optional"`
	Slug    string   `json:"slug" validate:"required,lte=16"`
	Name    string   `json:"name" validate:"required,lte=32"`
	Menus   []Menu   `json:"menus, optional" gorm:"many2many:au_role_menus;foreignKey:id;joinForeignKey:role_id;"`
	Permits []Permit `json:"permits, optional" gorm:"many2many:au_role_permits;foreignKey:id;joinForeignKey:role_id;"`
}

type RoleUpdateRequest struct {
	ID      int64    `json:"id"`
	Slug    string   `json:"slug" validate:"required,lte=16"`
	Name    string   `json:"name" validate:"required,lte=32"`
	Menus   []Menu   `json:"menus, optional" gorm:"many2many:au_role_menus;foreignKey:id;joinForeignKey:role_id;"`
	Permits []Permit `json:"permits, optional" gorm:"many2many:au_role_permits;foreignKey:id;joinForeignKey:role_id;"`
}

type RoleDeleteRequest struct {
	ID int64 `json:"id"`
}

type RoleQueryRequest struct {
	ID int64 `json:"id"`
}

type RoleQueryResponse struct {
	Role
}

type RoleQueryAllRequest struct {
	Name     string `json:"name, optional"`
	PageNum  int    `json:"pageNum, optional" gorm:"-"`
	PageSize int    `json:"pageSize, optional" gorm:"-"`
}

type RoleQueryAllResponse struct {
	List     []Role `json:"list"`
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
	RowCount int64  `json:"rowCount"`
}

type RoleDropDownItem struct {
	Id   int64  `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type OtpBindRequest struct {
	Account    string `json:"account" validate:"required,alphanumLength=6/12"`
	Password   string `json:"password,optional" validate:"omitempty,alphanumLength=8/16"`
	Regenerate string `json:"regenerate,optional"`
}

type OtpBindResponse struct {
	Secret string `json:"secret"`
	Qrcode string `json:"qrcode"`
}

type UserCurrency struct {
	UserAccount  string `json:"userAccount"`
	CurrencyCode string `json:"currencyCode"`
}

type UserCurrenciesQueryRequest struct {
	UserAccount string `json:"userAccount"`
}

type UserCurrenciesQueryResponse struct {
	List []UserCurrency `json:"list"`
}

type LoginRequest struct {
	Account  string `json:"account" validate:"required,alphanumLength=6/12"`
	Password string `json:"password" validate:"required,alphanumLength=8/16"`
	Otp      string `json:"otp,optional"`
}

type LoginResponse struct {
	Jwt JwtToken `json:"jwt"`
}

type JwtToken struct {
	AccessToken  string `json:"access_token"`
	AccessExpire int64  `json:"access_expire"`
	RefreshAfter int64  `json:"refresh_after"`
}

type OtpValidRequest struct {
	Otp string `json:"otp"`
}

type OtpValidResponse struct {
	Valid bool `json:"valid"`
}

type MerchantUser struct {
	ID            int64  `json:"id"`
	Merchant_code string `json:"merchantCode"`
	Account       string `json:"account"`
	Status        string `json:"status"`
	DisableDelete string `json:"disableDelete"`
	RegisteredAt  string `json:"registeredAt"`
	LastLogin_at  string `json:"lastLoginAt"`
	RoleNames     string `json:"roleNames"`
}

type MerchantUserQueryAllRequest struct {
	MerchantCode string `json:"merchantCode,optional"`
	Account      string `json:"account,optional"`
	RoleId       int64  `json:"roleId,optional"`
	Status       string `json:"status,optional"`
	PageNum      int    `json:"pageNum,optional" gorm:"-"`
	PageSize     int    `json:"pageSize,optional" gorm:"-"`
}

type MerchantUserCreateRequest struct {
	Account    string     `json:"account" validate:"alphanumLength=6/12"`
	Email      string     `json:"email" validate:"email"`
	Phone      string     `json:"phone, optional"`
	Roles      []Role     `json:"roles, optional" gorm:"many2many:au_user_roles;foreignKey:id;joinForeignKey:user_id;"`
	Merchants  []Merchant `json:"merchantsService, optional" gorm:"many2many:au_user_merchants;foreignKey:account;joinForeignKey:user_account;references:code;joinReferences:merchant_code;"`
	Currencies []Currency `json:"currencies, optional" gorm:"many2many:au_user_currencies;foreignKey:account;joinForeignKey:user_account;references:code;joinReferences:currency_code;"`
}

type MerchantUserUpdateRequest struct {
	ID         int64      `json:"id"`
	Account    string     `json:"account, optional" validate:"alphanumLength=6/12"`
	Email      string     `json:"email, optional" validate:"email"`
	Phone      string     `json:"phone, optional"`
	Status     string     `json:"status, optional"`
	Roles      []Role     `json:"roles, optional" gorm:"many2many:au_user_roles;foreignKey:id;joinForeignKey:user_id;"`
	Merchants  []Merchant `json:"merchantsService, optional" gorm:"many2many:au_user_merchants;foreignKey:account;joinForeignKey:user_account;references:code;joinReferences:merchant_code;"`
	Currencies []Currency `json:"currencies, optional" gorm:"many2many:au_user_currencies;foreignKey:account;joinForeignKey:user_account;references:code;joinReferences:currency_code;"`
}

type MerchantUserQueryAllResponse struct {
	List     []MerchantUser `json:"list"`
	PageNum  int            `json:"pageNum"`
	PageSize int            `json:"pageSize"`
	RowCount int64          `json:"rowCount"`
}

type MerchantUserResetPasswordRequest struct {
	ID int64 `json:"id"`
}

type MerchantUserUpdatePasswordRequest struct {
	ID              int64  `json:"id"`
	OldPassword     string `json:"oldPassword,optional" validate:"required" gorm:"-"`
	Password        string `json:"password" validate:"required,alphanumLength=8/16"`
	PasswordConfirm string `json:"passwordConfirm" validate:"eqfield=Password" gorm:"-"`
}

type Timezonex struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type TimezonexCreateRequest struct {
	Code string `json:"code" validate: "required, lte=2"`
	Name string `json:"name" validate: "required, lte=64"`
}

type TimezonexUpdateRequest struct {
	ID   int64  `json:"id"`
	Code string `json:"code" validate: "required, lte=2"`
	Name string `json:"name" validate: "required, lte=64"`
}

type TimezonexDeleteRequest struct {
	ID int64 `json:"id"`
}

type TimezonexQueryRequest struct {
	ID int64 `json:"id"`
}

type TimezonexQueryResponse struct {
	Timezonex
}

type TimezonexQueryAllRequest struct {
	Code     string `json:"code, optional"`
	Name     string `json:"name, optional"`
	PageNum  int    `json:"pageNum" gorm:"-"`
	PageSize int    `json:"pageSize" gorm:"-"`
}

type TimezonexQueryAllResponse struct {
	List     []Timezonex `json:"list"`
	PageNum  int         `json:"pageNum"`
	PageSize int         `json:"pageSize"`
	RowCount int64       `json:"rowCount"`
}

type Currency struct {
	ID   int64  `json:"id, optional"`
	Code string `json:"code, optional"`
	Name string `json:"name, optional"`
}

type CurrencyCreateRequest struct {
	Code string `json:"code" valiate:"required,lte=4"`
	Name string `json:"name" valiate:"required,lte=16"`
}

type CurrencyUpdateRequest struct {
	ID   int64  `json:"id"`
	Code string `json:"code" valiate:"required,lte=4"`
	Name string `json:"name" valiate:"required,lte=16"`
}

type CurrencyDeleteRequest struct {
	ID int64 `json:"id"`
}

type CurrencyQueryRequest struct {
	ID int64 `json:"id"`
}

type CurrencyQueryResponse struct {
	Currency
}

type CurrencyQueryAllRequest struct {
	Code     string `json:"code, optional"`
	Name     string `json:"name, optional"`
	PageNum  int    `json:"pageNum, optional" gorm:"-"`
	PageSize int    `json:"pageSize, optional" gorm:"-"`
}

type CurrencyQueryAllResponse struct {
	List     []Currency `json:"list"`
	PageNum  int        `json:"pageNum"`
	PageSize int        `json:"pageSize"`
	RowCount int64      `json:"rowCount"`
}

type LanguageCreateRequest struct {
	Code string `json:"code" valiate:"required,lte=2"`
	Name string `json:"name" valiate:"required,lte=16"`
}

type LanguageUpdateRequest struct {
	ID   string `json:"id"`
	Code string `json:"code" valiate:"required,lte=2"`
	Name string `json:"name" valiate:"required,lte=16"`
}

type LanguageQueryRequest struct {
	ID int64 `json:"id"`
}

type LanguageDeleteRequest struct {
	ID int64 `json:"id"`
}

type Language struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type LanguageQueryResponse struct {
	Language
}

type LanguageQueryAllRequest struct {
	Code     string `json:"code ,optional"`
	Name     string `json:"name,optional"`
	PageNum  int    `json:"pageNum" gorm:"-"` //gorm:"-" 忽略這參數檢查是否有給
	PageSize int    `json:"pageSize" gorm:"-"`
}

type LanguageQueryAllResponse struct {
	List     []Language `json:"list"`
	PageNum  int        `json:"pageNum"`
	PageSize int        `json:"pageSize"`
	RowCount int64      `json:"rowCount"`
}

type BankBlockAccount struct {
	ID          int64  `json:"id"`
	BankAccount string `json:"bankAccount"`
	BankNo      string `json:"bankNo"`
	Memo        string `json:"memo"`
	CreatedAt   string `json:"createdAt"`
	Name        string `json:"name"`
}

type BankBlockAccountCreateRequest struct {
	BankAccount string `json:"bankAccount" validate: "required"`
	Name        string `json:"name" validate: "required"`
	Memo        string `json:"memo"`
}

type BankBlockAccountUpdateRequest struct {
	ID          int64  `json:"id"`
	BankAccount string `json:"bankAccount" validate: "required"`
	Name        string `json:"name" validate: "required"`
	Memo        string `json:"memo"`
}

type BankBlockAccountDeleteRequest struct {
	ID int64 `json:"id"`
}

type BankBlockAccountQueryRequest struct {
	ID int64 `json:"id"`
}

type BankBlockAccountQueryResponse struct {
	BankBlockAccount
}

type BankBlockAccountQueryAllRequest struct {
	BankAccount string `json:"bankAccount, optional"`
	Name        string `json:"name, optional"`
	StartAt     string `json:"startAt, optional"`
	EndAt       string `json:"endAt, optional"`
	PageNum     int    `json:"pageNum" gorm:"-"`
	PageSize    int    `json:"pageSize" gorm:"-"`
}

type BankBlockAccountQueryAllResponse struct {
	List     []BankBlockAccount `json:"list"`
	PageNum  int                `json:"pageNum"`
	PageSize int                `json:"pageSize"`
	RowCount int64              `json:"rowCount"`
}

type Bank struct {
	ID           int64         `json:"id"`
	BankNo       string        `json:"bankNo"`
	BankName     string        `json:"bankName"`
	BankNameEn   string        `json:"bankNameEn"`
	Abbr         string        `json:"abbr"`
	BranchNo     string        `json:"branchNo"`
	BranchName   string        `json:"branchName"`
	City         string        `json:"city"`
	Province     string        `json:"province"`
	CurrencyCode string        `json:"currencyCode"`
	Status       string        `json:"status"`
	ChannelDatas []ChannelData `json:"channelDatas, optional" gorm:"many2many:ch_channel_banks:foreignKey:bankNo;joinForeignKey:bank_no;references:code;joinReferences:channel_code"`
}

type BankCreateRequest struct {
	BankNo       string        `json:"bankNo, optional"`
	BankName     string        `json:"bankName, valiate: "required""`
	BankNameEn   string        `json:"bankNameEn" valiate: "required"`
	Abbr         string        `json:"abbr" valiate: "required"`
	BranchNo     string        `json:"branchNo, optional "`
	BranchName   string        `json:"branchName, optional"`
	City         string        `json:"city, optional"`
	Province     string        `json:"province, optional"`
	CurrencyCode string        `json:"currencyCode" valiate: "required"`
	Status       string        `json:"status" valiate: "required"`
	ChannelDatas []ChannelData `json:"channelDatas, optional" gorm:"many2many:ch_channel_banks:foreignKey:bankNo;joinForeignKey:bank_no;references:code;joinReferences:channel_code"`
}

type BankUpdateRequest struct {
	ID           int64         `json:"id, optional"`
	BankNo       string        `json:"bankNo" valiate: "required"`
	BankName     string        `json:"bankName, optional"`
	BankNameEn   string        `json:"bankNameEn, optional"`
	Abbr         string        `json:"abbr, optional"`
	BranchNo     string        `json:"branchNo, optional"`
	BranchName   string        `json:"branchName, optional"`
	City         string        `json:"city, optional"`
	Province     string        `json:"province, optional"`
	CurrencyCode string        `json:"currencyCode" valiate: "required"`
	Status       string        `json:"status" valiate: "required"`
	ChannelDatas []ChannelData `json:"channelDatas, optional" gorm:"many2many:ch_channel_banks:foreignKey:bankNo;joinForeignKey:bank_no;references:code;joinReferences:channel_code"`
}

type BankDeleteRequest struct {
	ID int64 `json:"id"`
}

type BankQueryRequest struct {
	ID int64 `json:"id"`
}

type BankQueryResponse struct {
	Bank
}

type BankQueryAllRequest struct {
	CurrencyCode string `json:"currencyCode, optional"`
	BankNo       string `json:"bankNo, optional"`
	BankName     string `json:"bankName, optional"`
	PageNum      int    `json:"pageNum, optional" gorm:"-"`
	PageSize     int    `json:"pageSize, optional" gorm:"-"`
}

type BankQueryAllResponse struct {
	List     []Bank `json:"list"`
	PageNum  int    `json:"pageNum" gorm:"-"`
	PageSize int    `json:"pageSize" gorm:"-"`
	RowCount int64  `json:"rowCount"`
}

type Merchant struct {
	ID                   int64              `json:"id, optional"`
	Code                 string             `json:"code, optional"`
	AccountName          string             `json:"accountName, optional"`
	AgentLayerCode       string             `json:"agentLayerCode, optional"`
	AgentParentCode      string             `json:"agentParentCode, optional"`
	ScrectKey            string             `json:"-, optional"`
	AccessToken          string             `json:"-, optional"`
	Status               string             `json:"status, optional"`
	AgentStatus          string             `json:"agentStatus, optional"`
	Contact              MerchantContact    `json:"contact, optional"`
	BizInfo              MerchantBizInfo    `json:"bizInfo, optional"`
	BoIP                 string             `json:"boIp, optional"`
	ApiIP                string             `json:"apiIp, optional"`
	IsWithdraw           string             `json:"isWithdraw, optional"`
	WithdrawPassword     string             `json:"-, optional"`
	LoginValidatedType   string             `json:"loginValidatedType, optional"`
	PayingValidatedType  string             `json:"payingValidatedType, optional"`
	ApiCodeType          string             `json:"apiCodeType, optional"`
	BillLadingType       string             `json:"billLadingType, optional"`
	Lang                 string             `json:"lang, optional"`
	RegisteredAt         int64              `json:"registeredAt, optional"`
	RegisteredAtS        string             `json:"registeredAtS, optional" gorm:"-"`
	RateCheck            string             `json:"rateCheck, optional"`
	UnpaidNotifyInterval int64              `json:"unpaidNotifyInterval, optional"`
	UnpaidNotifyNum      int64              `json:"unpaidNotifyNum, optional"`
	MerchantBalances     []MerchantBalance  `json:"merchantBalances, optional" gorm:"foreignKey:MerchantCode;references:Code"`
	Users                []User             `json:"users, optional, optional" gorm:"many2many:au_user_merchants;foreignKey:code;joinForeignKey:merchant_code;references:account;joinReferences:user_account;"`
	MerchantCurrencies   []MerchantCurrency `json:"merchantCurrencies, optional" gorm:"foreignKey:MerchantCode;references:Code"`
}

type MerchantContact struct {
	Phone                 string `json:"phone, optional"`
	Email                 string `json:"email, optional"`
	CommunicationSoftware string `json:"communicationSoftware, optional"`
	CommunicationNickname string `json:"communicationNickname, optional"`
	GroupName             string `json:"groupName, optional"`
	GroupID               string `json:"groupId, optional"`
}

type MerchantBizInfo struct {
	CompanyName      string `json:"companyName, optional"`
	OperatingWebsite string `json:"operatingWebsite, optional"`
	TestAccount      string `json:"testAccount, optional"`
	TestPassword     string `json:"testPassword, optional"`
	Remark           string `json:"remark, optional"`
}

type MerchantCreateRequest struct {
	Account             string             `json:"account" gorm:"-"`
	Code                string             `json:"code, optional"`
	AccountName         string             `json:"accountName, optional"`
	AgentLayerCode      string             `json:"agentLayerCode, optional"`
	AgentParentCode     string             `json:"agentParentCode, optional"`
	ScrectKey           string             `json:"screctKey, optional"`
	AccessToken         string             `json:"accessToken, optional"`
	Status              string             `json:"status, optional"`
	AgentStatus         string             `json:"agentStatus, optional"`
	Contact             MerchantContact    `json:"contact"`
	BizInfo             MerchantBizInfo    `json:"bizInfo, optional"`
	BoIP                string             `json:"boIp, optional"`
	ApiIP               string             `json:"apiIp, optional"`
	IsWithdraw          string             `json:"isWithdraw, optional"`
	WithdrawPassword    string             `json:"withdrawPassword, optional"`
	LoginValidatedType  string             `json:"loginValidatedType, optional"`
	PayingValidatedType string             `json:"payingValidatedType, optional"`
	ApiCodeType         string             `json:"apiCodeType, optional"`
	BillLadingType      string             `json:"billLadingType, optional"`
	Lang                string             `json:"lang, optional"`
	RegisteredAt        int64              `json:"registeredAt, optional"`
	Users               []User             `json:"users, optional" gorm:"many2many:au_user_merchants;foreignKey:code;joinForeignKey:merchant_code;references:account;joinReferences:user_account;"`
	MerchantCurrencies  []MerchantCurrency `json:"merchantCurrencies, optional" gorm:"foreignKey:MerchantCode;references:Code"`
}

type MerchantUpdateRequest struct {
	ID                  int64           `json:"id"`
	Code                string          `json:"code, optional"`
	AgentLayerCode      string          `json:"agentLayerCode, optional"`
	AgentParentCode     string          `json:"agentParentCode, optional"`
	ScrectKey           string          `json:"screctKey, optional"`
	AccessToken         string          `json:"accessToken, optional"`
	Status              string          `json:"status, optional"`
	AgentStatus         string          `json:"agentStatus, optional"`
	Contact             MerchantContact `json:"contact, optional"`
	BizInfo             MerchantBizInfo `json:"bizInfo, optional"`
	BoIP                string          `json:"boIp, optional"`
	ApiIP               string          `json:"apiIp, optional"`
	IsWithdraw          string          `json:"isWithdraw, optional"`
	WithdrawPassword    string          `json:"withdrawPassword, optional"`
	LoginValidatedType  string          `json:"loginValidatedType, optional"`
	PayingValidatedType string          `json:"payingValidatedType, optional"`
	ApiCodeType         string          `json:"apiCodeType, optional"`
	BillLadingType      string          `json:"billLadingType, optional"`
	Lang                string          `json:"lang, optional"`
	RegisteredAt        int64           `json:"registeredAt, optional"`
}

type MerchantDeleteRequest struct {
	ID int64 `json:"id"`
}

type MerchantQueryRequest struct {
	ID int64 `json:"id"`
}

type MerchantQueryResponse struct {
	Merchant
}

type MerchantQueryAllRequest struct {
	Code           string `json:"code, optional"`
	AgentLayerCode string `json:"agentLayerCode, optional"`
	AccountName    string `json:"accountName, optional"`
	PageNum        int    `json:"pageNum, optional" gorm:"-"`
	PageSize       int    `json:"pageSize, optional" gorm:"-"`
}

type MerchantQueryAllResponse struct {
	List     []Merchant `json:"list"`
	PageNum  int        `json:"pageNum"`
	PageSize int        `json:"pageSize"`
	RowCount int64      `json:"rowCount"`
}

type MerchantListView struct {
	ID                  int64   `json:"id, optional"`
	Code                string  `json:"code, optional"`
	AgentLayerCode      string  `json:"agentLayerCode, optional"`
	Status              string  `json:"status, optional"`
	CreatedAt           string  `json:"createdAt, optional"`
	AgentStatus         string  `json:"agentStatus, optional"`
	CurrencyCode        string  `json:"currencyCode, optional" validate:"required"`
	XfBalance           float64 `json:"xfBalance, optional"`
	DfBalance           float64 `json:"dfBalance, optional"`
	FrozenAmount        float64 `json:"frozenAmount, optional"`
	Commission          float64 `json:"commission, optional"`
	WithdrawHandlingFee float64 `json:"withdrawHandlingFee, optional"`
	MinWithdrawCharge   float64 `json:"minWithdrawCharge"`
	MaxWithdrawCharge   float64 `json:"maxWithdrawCharge"`
}

type MerchantQueryListViewRequest struct {
	Code                     string   `json:"code, optional"`
	AgentLayerCode           string   `json:"agentLayerCode, optional"`
	BalanceCurrencyCode      string   `json:"balanceCurrencyCode, optional"`
	Currencies               []string `json:"currencies, optional"`
	UserID                   int64    `json:"userId, optional"`
	MerchantCurrenciesStatus string   `json:"merchantCurrenciesStatus, optional"`
	PageNum                  int      `json:"pageNum, optional" gorm:"-"`
	PageSize                 int      `json:"pageSize, optional" gorm:"-"`
}

type MerchantQueryListViewResponse struct {
	List     []MerchantListView `json:"list"`
	PageNum  int                `json:"pageNum"`
	PageSize int                `json:"pageSize"`
	RowCount int64              `json:"rowCount"`
}

type MerchantQueryDistinctCode struct {
	CurrencyCode string `json:"currencyCode, optional"`
	Status       string `json:"status, optional"`
}

type MerchantUpdateStatusRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type MerchantUpdateAgentStatusRequest struct {
	ID          int64  `json:"id"`
	AgentStatus string `json:"agentStatus"`
}

type MerchantUpdateBillLadingTypeRequest struct {
	ID             int64  `json:"id"`
	BillLadingType string `json:"billLadingType"`
}

type MerchantResetPasswordRequest struct {
	Name string `json:"name"`
}

type MerchantResetScrectKeyRequest struct {
	ID int64 `json:"id"`
}

type MerchantGetScrectKeyRequest struct {
	ID int64 `json:"id"`
}

type MerchantGetScrectKeyResponse struct {
	ScrectKey string `json:"screctKey"`
}

type MerchantTransferParentAgentRequest struct {
	Code            string `json:"code"`
	AgentParentCode string `json:"agentParentCode"`
}

type UpdateWithdrawPasswordRequest struct {
	ID               int64  `json:"id"`
	WithdrawPassword string `json:"withdrawPassword" validate:"required,alphanumLength=8/16"`
}

type ResetWithdrawPasswordRequest struct {
	ID int64 `json:"id"`
}

type MerchantBalance struct {
	ID           int64   `json:"id, optional"`
	MerchantCode string  `json:"merchantCode, optional"`
	CurrencyCode string  `json:"currencyCode, optional"`
	BalanceType  string  `json:"balanceType, optional"  validate:"required"`
	Balance      float64 `json:"balance"`
	FrozenAmount float64 `json:"frozenAmount, optional"`
}

type MerchantPtBalance struct {
	ID           int64   `json:"id, optional"`
	MerchantCode string  `json:"merchantCode"`
	CurrencyCode string  `json:"currencyCode"`
	Name         string  `json:"name"`
	Balance      float64 `json:"balance"`
	Remark       string  `json:"remark"`
}

type MerchantBalanceCreateRequest struct {
	ID           int64   `json:"id, optional"`
	MerchantCode string  `json:"merchantCode, optional" validate:"required"`
	CurrencyCode string  `json:"currencyCode, optional" validate:"required"`
	BalanceType  string  `json:"balanceType, optional"  validate:"required"`
	Balance      float64 `json:"balance"`
	FrozenAmount float64 `json:"frozenAmount, optional"`
	Commission   float64 `json:"commission, optional"`
}

type MerchantBalanceUpdateRequest struct {
	MerchantCode string  `json:"merchantCode"`
	CurrencyCode string  `json:"currencyCode"`
	BalanceType  string  `json:"balanceType"`
	Comment      string  `json:"comment"`
	Amount       float64 `json:"amount"`
}

type MerchantBalanceDeleteRequest struct {
	ID int64 `json:"id"`
}

type MerchantBalanceQueryRequest struct {
	ID int64 `json:"id"`
}

type MerchantBalanceQueryResponse struct {
	MerchantBalance
}

type MerchantBalanceQueryAllRequest struct {
	MerchantCode string `json:"merchantCode, optional"`
	CurrencyCode string `json:"currencyCode, optional"`
	PageNum      int    `json:"pageNum, optional" gorm:"-"`
	PageSize     int    `json:"pageSize, optional" gorm:"-"`
}

type MerchantBalanceQueryAllResponse struct {
	List     []MerchantBalance `json:"list"`
	PageNum  int               `json:"pageNum"`
	PageSize int               `json:"pageSize"`
	RowCount int64             `json:"rowCount"`
}

type MerchantChannelRate struct {
	ID                  int64   `json:"id"`
	MerchantCode        string  `json:"merchantCode"`
	ChannelPayTypesCode string  `json:"channelPayTypesCode"`
	ChannelCode         string  `json:"channelCode"`
	PayTypeCode         string  `json:"payTypeCode"`
	Designation         string  `json:"designation"`
	DesignationNo       string  `json:"designationNo, optional"`
	Fee                 float64 `json:"fee"`
	HandlingFee         float64 `json:"handlingFee"`
	Status              string  `json:"status"`
}

type MerchantChannelRateConfigureRequest struct {
	ID                  int64   `json:"id, optional"`
	MerchantCode        string  `json:"merchantCode"`
	ChannelPayTypesCode string  `json:"channelPayTypesCode"`
	ChannelCode         string  `json:"channelCode"`
	PayTypeCode         string  `json:"payTypeCode"`
	Designation         string  `json:"designation"`
	DesignationNo       string  `json:"designationNo, optional"`
	Fee                 float64 `json:"fee"`
	HandlingFee         float64 `json:"handlingFee"`
	Status              string  `json:"status, optional"`
}

type MerchantConfigureRate struct {
	IsConfigure                 bool    `json:"isConfigure"`
	ChnName                     string  `json:"chnName"`
	ChnCode                     string  `json:"chnCode"`
	ChnIsNzPre                  string  `json:"chnIsNzPre"`
	ChnIsProxy                  string  `json:"chnIsProxy"`
	ChnStatus                   string  `json:"chnStatus"`
	CurrencyCode                string  `json:"currencyCode"`
	PayTypeCode                 string  `json:"payTypeCode"`
	ChnPayTypeCode              string  `json:"chnPayTypeCode"`
	PayTypeFee                  float64 `json:"payTypeFee"`
	PayTypeHandlingFee          float64 `json:"payTypeHandlingFee"`
	ChnPayTypeMaxInternalCharge float64 `json:"chnPayTypeMaxInternalCharge"`
	ChnPayTypeDailyTxLimit      float64 `json:"chnPayTypeDailyTxLimit"`
	ChnPayTypeSingleMinCharge   float64 `json:"chnPayTypeSingleMinCharge"`
	ChnPayTypeSingleMaxCharge   float64 `json:"chnPayTypeSingleMaxCharge"`
	ChnPayTypeBillDate          int     `json:"chnPayTypeBillDate"`
	ChnPayTypeStatus            string  `json:"chnPayTypeStatus"`
	MerCode                     string  `json:"merCode"`
	MerBillLadingType           string  `json:"merBillLadingType"`
	MerStatus                   string  `json:"merStatus"`
	MerChnRateId                int64   `json:"merChnRateId"`
	MerChnRateFee               float64 `json:"merChnRateFee"`
	MerChnRateHandlingFee       float64 `json:"merChnRateHandlingFee"`
	MerChnRateDesignation       string  `json:"merChnRateDesignation"`
	MerChnRateDesignationNo     string  `json:"merChnRateDesignationNo"`
	MerChnRateStatus            string  `json:"merChnRateStatus"`
	ParentMerCode               string  `json:"parentMerCode"`
	ParentMerChnRateStatus      string  `json:"parentMerChnRateStatus"`
}

type MerchantConfigureRateListRequest struct {
	CurrencyCode string `json:"currencyCode"`
	MerchantCode string `json:"merchantCode"`
	PayTypeCode  string `json:"payTypeCode"`
	PageNum      int    `json:"pageNum, optional" gorm:"-"`
	PageSize     int    `json:"pageSize, optional" gorm:"-"`
}

type MerchantConfigureRateListResponse struct {
	List []MerchantConfigureRate `json:"list"`
}

type MerchantRateListView struct {
	ChnName                     string  `json:"chnName"`
	ChnCode                     string  `json:"chnCode"`
	CurrencyCode                string  `json:"currencyCode"`
	CurrencyName                string  `json:"currencyName"`
	ChnStatus                   string  `json:"chnStatus"`
	PayTypeCode                 string  `json:"payTypeCode"`
	PayTypeName                 string  `json:"payTypeName"`
	ChnPayTypeCode              string  `json:"chnPayTypeCode"`
	ChnPayTypeFee               float64 `json:"chnPayTypeFee"`
	ChnPayTypeHandlingFee       float64 `json:"chnPayTypeHandlingFee"`
	ChnPayTypeMaxInternalCharge float64 `json:"chnPayTypeMaxInternalCharge"`
	ChnPayTypeDailyTxLimit      float64 `json:"chnPayTypeDailyTxLimit"`
	ChnPayTypeSingleMinCharge   float64 `json:"chnPayTypeSingleMinCharge"`
	ChnPayTypeSingleMaxCharge   float64 `json:"chnPayTypeSingleMaxCharge"`
	ChnPayTypeBillDate          int     `json:"chnPayTypeBillDate"`
	ChnPayTypeStatus            string  `json:"chnPayTypeStatus"`
	MerCode                     string  `json:"merCode"`
	MerStatus                   string  `json:"merStatus"`
	MerChnRateFee               float64 `json:"merChnRateFee"`
	MerChnRateHandlingFee       float64 `json:"merChnRateHandlingFee"`
	MerChnRateDesignation       string  `json:"merChnRateDesignation"`
	MerChnRateStatus            string  `json:"merChnRateStatus"`
	ParentMerCode               string  `json:"parentMerCode"`
	ParentMerChnRateFee         float64 `json:"parentMerChnRateFee"`
	ParentMerChnRateHandlingFee float64 `json:"parentMerChnRateHandlingFee"`
	ParentMerChnRateStatus      string  `json:"parentMerChnRateStatus"`
}

type MerchantOrderRateListView struct {
	ChannelPayTypesCode string  `json:"channelPayTypesCode"`
	PayTypeCode         string  `json:"payTypeCode"`
	MerHandlingFee      float64 `json:"merHandlingFee"`
	MerFee              float64 `json:"merFee"`
	Designation         string  `json:"designation"`
	DesignationNo       string  `json:"designationNo"`
	ChannelCode         string  `json:"channelCode"`
	ChannelName         string  `json:"channelName"`
	CurrencyCode        string  `json:"currencyCode"`
	MaxInternalCharge   float64 `json:"maxInternalCharge"`
	SingleMinCharge     float64 `json:"singleMinCharge"`
	SingleMaxCharge     float64 `json:"singleMaxCharge"`
	MerchantCode        string  `json:"merchantCode"`
	MerchnrateStatus    string  `json:"merchnrateStatus"`
	ChnStatus           string  `json:"chnStatus"`
	ChnIsProxy          string  `json:"chnIsProxy"`
	CptStatus           string  `json:"chnpayTypeStatus"`
	CptFee              float64 `json:"cptFee"`
	CptHandlingFee      float64 `json:"cptHandlingFee"`
	ApiUrl              string  `json:"apiUrl"`
	ChannelPort         string  `json:"channelPort"`
	CptIsRate           string  `json:"cptIsRate"`
	Device              string  `json:"device"`
	FixedAmount         string  `json:"fixedAmount"`
	PayTypeName         string  `json:"payTypeName"`
	PayTypeImgUrl       string  `json:"payTypeImgUrl"`
}

type MerchantQueryRateListViewRequest struct {
	CurrencyCode       string `json:"currencyCode, optional"`
	MerchantCode       string `json:"merchantCode, optional"`
	ParentMerchantCode string `json:"parentMerchantCode, optional"`
	PayTypeCode        string `json:"payTypeCode, optional"`
	PayTypeName        string `json:"payTypeName, optional"`
	ChannelName        string `json:"channelName, optional"`
	ChannelPayTypeCode string `json:"channelPayTypeCode, optional"`
	PageNum            int    `json:"pageNum, optional" gorm:"-"`
	PageSize           int    `json:"pageSize, optional" gorm:"-"`
}

type MerchantQueryRateListViewResponse struct {
	List     []MerchantRateListView `json:"list"`
	PageNum  int                    `json:"pageNum, optional" gorm:"-"`
	PageSize int                    `json:"pageSize, optional" gorm:"-"`
	RowCount int64                  `json:"rowCount"`
}

type MerchantRateGetCodeDropDownListRequest struct {
	CurrencyCode string `json:"currencyCode, optional"`
}

type MerchantRateGetParentCodeDropDownListRequest struct {
	CurrencyCode string `json:"currencyCode, optional"`
}

type MerchantCurrency struct {
	ID                  int64   `json:"id, optional"`
	MerchantCode        string  `json:"merchantCode, optional"`
	CurrencyCode        string  `json:"currencyCode, optional"`
	WithdrawHandlingFee float64 `json:"withdrawHandlingFee, optional"`
	MinWithdrawCharge   float64 `json:"minWithdrawCharge, optional"`
	MaxWithdrawCharge   float64 `json:"maxWithdrawCharge, optional"`
	Status              string  `json:"status, optional"`
	IsDisplayPtBalance  string  `json:"isDisplayPtBalance, optional"`
}

type MerchantCurrencyCreateRequest struct {
	MerchantCode        string  `json:"merchantCode, optional"`
	CurrencyCode        string  `json:"currencyCode, optional"`
	WithdrawHandlingFee float64 `json:"withdrawHandlingFee, optional"`
	MinWithdrawCharge   float64 `json:"minWithdrawCharge, optional"`
	MaxWithdrawCharge   float64 `json:"maxWithdrawCharge, optional"`
	Status              string  `json:"status, optional"`
}

type MerchantCurrencyUpdateRequest struct {
	ID                  int64   `json:"id, optional"`
	MerchantCode        string  `json:"merchantCode, optional"`
	CurrencyCode        string  `json:"currencyCode, optional"`
	WithdrawHandlingFee float64 `json:"withdrawHandlingFee, optional"`
	MinWithdrawCharge   float64 `json:"minWithdrawCharge, optional"`
	MaxWithdrawCharge   float64 `json:"maxWithdrawCharge, optional"`
	Status              string  `json:"status, optional"`
}

type MerchantCurrencyQueryRequest struct {
	ID int64 `json:"id"`
}

type MerchantCurrencyQueryResponse struct {
	MerchantCurrency
}

type MerchantCurrencyQueryAllRequest struct {
	MerchantCode string `json:"merchantCode, optional"`
	CurrencyCode string `json:"currencyCode, optional"`
	Status       string `json:"status, optional"`
	PageNum      int    `json:"pageNum, optional" gorm:"-"`
	PageSize     int    `json:"pageSize, optional" gorm:"-"`
}

type MerchantCurrencyQueryAllResponse struct {
	List     []MerchantCurrency `json:"list"`
	PageNum  int                `json:"pageNum"`
	PageSize int                `json:"pageSize"`
	RowCount int64              `json:"rowCount"`
}

type MerchantGetAvailableCurrencyRequest struct {
	MerchantCode string `json:"merchantCode"`
}

type MerchantUpdateCurrenciesRequest struct {
	MerchantCode string `json:"merchantCode"`
}

type MerchantBindBank struct {
	ID           int64  `json:"id"`
	MerchantCode string `json:"merchantCode"`
	CurrencyCode string `json:"currencyCode"`
	Name         string `json:"name"`
	BankAccount  string `json:"bankAccount"`
	BankNo       string `json:"bankNo"`
	BankName     string `json:"bankName"`
	Province     string `json:"province"`
	City         string `json:"city"`
}

type MerchantBindBankCreateRequest struct {
	MerchantCode string `json:"merchantCode, optional"`
	CurrencyCode string `json:"currencyCode"`
	Name         string `json:"name"`
	BankAccount  string `json:"bankAccount"`
	BankNo       string `json:"bankNo"`
	BankName     string `json:"bankName"`
	Province     string `json:"province"`
	City         string `json:"city"`
}

type MerchantBindBankUpdateRequest struct {
	ID           int64  `json:"id"`
	CurrencyCode string `json:"currencyCode, optional"`
	Name         string `json:"name, optional"`
	BankAccount  string `json:"bankAccount, optional"`
	BankNo       string `json:"bankNo, optional"`
	BankName     string `json:"bankName, optional"`
	Province     string `json:"province, optional"`
	City         string `json:"city, optional"`
}

type MerchantBindBankDeleteRequest struct {
	ID int64 `json:"id"`
}

type MerchantBindBankQueryRequest struct {
	ID int64 `json:"id"`
}

type MerchantBindBankQueryResponse struct {
	MerchantBindBank
}

type MerchantBindBankQueryAllRequest struct {
	CurrencyCode string `json:"currencyCode, optional"`
	PageNum      int    `json:"pageNum" gorm:"-"`
	PageSize     int    `json:"pageSize" gorm:"-"`
}

type MerchantBindBankQueryAllResponse struct {
	List     []MerchantBindBank `json:"list"`
	PageNum  int                `json:"pageNum"`
	PageSize int                `json:"pageSize"`
	RowCount int64              `json:"rowCount"`
}

type MerchantBalanceRecord struct {
	ID                int64       `json:"id, optional"`
	MerchantBalanceId int64       `json:"merchantBalanceId"`
	MerchantCode      string      `json:"merchantCode, optional"`
	CurrencyCode      string      `json:"currencyCode, optional"`
	OrderNo           string      `json:"orderNo"`
	OrderType         string      `json:"orderType"`
	ChannelCode       string      `json:"channelCode"`
	PayTypeCode       string      `json:"payTypeCode"`
	TransactionType   string      `json:"transactionType"`
	BalanceType       string      `json:"balanceType, optional"`
	BeforeBalance     float64     `json:"beforeBalance"`
	TransferAmount    float64     `json:"transferAmount"`
	AfterBalance      float64     `json:"afterBalance"`
	Comment           string      `json:"comment"`
	CreatedBy         string      `json:"createdBy"`
	CreatedAt         string      `json:"createdAt"`
	PayTypeData       PayType     `json:"payTypeData, optional" gorm:"foreignKey:Code;references:PayTypeCode"`
	ChannelData       ChannelData `json:"channelData, optional" gorm:"foreignKey:Code;references:ChannelCode"`
}

type MerchantBalanceRecordQueryAllRequest struct {
	MerchantBalanceId int64  `json:"merchantBalanceId, optional"`
	MerchantCode      string `json:"merchantCode, optional"`
	CurrencyCode      string `json:"currencyCode, optional"`
	OrderNo           string `json:"orderNo, optional"`
	OrderType         string `json:"orderType, optional"`
	TransactionType   string `json:"transactionType, optional"`
	BalanceType       string `json:"balanceType, optional"`
	PayTypeCode       string `json:"payTypeCode, optional"`
	StartAt           string `json:"startAt"`
	EndAt             string `json:"endAt"`
	PageNum           int    `json:"pageNum, optional" gorm:"-"`
	PageSize          int    `json:"pageSize, optional" gorm:"-"`
}

type MerchantBalanceRecordQueryAllResponse struct {
	List     []MerchantBalanceRecord `json:"list"`
	PageNum  int                     `json:"pageNum"`
	PageSize int                     `json:"pageSize"`
	RowCount int64                   `json:"rowCount"`
}

type MerchantFrozenRecord struct {
	ID                int64   `json:"id, optional"`
	MerchantBalanceId int64   `json:"merchantBalanceId"`
	MerchantCode      string  `json:"merchantCode, optional"`
	CurrencyCode      string  `json:"currencyCode, optional"`
	OrderNo           string  `json:"orderNo"`
	OrderType         string  `json:"orderType"`
	TransactionType   string  `json:"transactionType"`
	BeforeFrozen      float64 `json:"beforeFrozen"`
	FrozenAmount      float64 `json:"frozenAmount"`
	AfterFrozen       float64 `json:"afterFrozen"`
	Comment           string  `json:"comment"`
	CreatedBy         string  `json:"createdBy"`
}

type MerchantFrozenRecordQueryAllRequest struct {
	MerchantBalanceId int64  `json:"merchantBalanceId, optional"`
	MerchantCode      string `json:"merchantCode, optional"`
	CurrencyCode      string `json:"currencyCode, optional"`
	OrderNo           string `json:"orderNo, optional"`
	OrderType         string `json:"orderType, optional"`
	TransactionType   string `json:"transactionType, optional"`
	PageNum           int    `json:"pageNum, optional" gorm:"-"`
	PageSize          int    `json:"pageSize, optional" gorm:"-"`
}

type MerchantFrozenRecordQueryAllResponse struct {
	List     []MerchantFrozenRecord `json:"list"`
	PageNum  int                    `json:"pageNum"`
	PageSize int                    `json:"pageSize"`
	RowCount int64                  `json:"rowCount"`
}

type Order struct {
	ID                      int64   `json:"id"`
	Type                    string  `json:"type"` //收支方式  (代付 DF 支付 ZF 下發 XF 內充 NC)
	MerchantCode            string  `json:"merchantCode"`
	TransAt                 string  `json:"transAt"`
	OrderNo                 string  `json:"orderNo"`
	BalanceType             string  `json:"balanceType"`
	BeforeBalance           float64 `json:"beforeBalance"`
	TransferAmount          float64 `json:"transferAmount"`
	Balance                 float64 `json:"balance"`
	FrozenAmount            float64 `json:"frozenAmount"`
	Status                  string  `json:"status"` //訂單狀態(0:待處理 1:處理中 2:交易中 20:成功 30:失敗 31:凍結)
	IsLock                  string  `json:"isLock"`
	RepaymentStatus         string  `json:"repaymentStatus"` //还款状态：(0：不需还款、1:待还款、2：还款成功、3：还款失败)
	Memo                    string  `json:"memo"`
	ErrorType               string  `json:"errorType, optional"`
	ErrorNote               string  `json:"errorNote, optional"`
	ChannelCode             string  `json:"channelCode"`
	ChannelPayTypesCode     string  `json:"channelPayTypesCode"`
	PayTypeCode             string  `json:"payTypeCode"`
	Fee                     float64 `json:"fee"`
	HandlingFee             float64 `json:"handlingFee"`
	InternalChargeOrderPath string  `json:"internalChargeOrderPath"`
	CurrencyCode            string  `json:"currencyCode"`
	MerchantBankAccount     string  `json:"merchantBankAccount"`  //商戶銀行帳號
	MerchantBankNo          string  `json:"merchantBankNo"`       //商戶銀行代碼
	MerchantBankName        string  `json:"merchantBankName"`     //商戶姓名
	MerchantBankBranch      string  `json:"merchantBankBranch"`   //商戶銀行分行名
	MerchantBankProvince    string  `json:"merchantBankProvince"` //商戶開戶縣市名
	MerchantBankCity        string  `json:"merchantBankCity"`     //商戶開戶縣市名
	MerchantAccountName     string  `json:"merchantAccountName"`  //開戶行名(代付、下發)
	ChannelBankAccount      string  `json:"channelBankAccount"`   //渠道銀行帳號
	ChannelBankNo           string  `json:"channelBankNo"`        //渠道銀行代碼
	ChannelBankName         string  `json:"channelBankName"`      //渠道银行名称
	ChannelAccountName      string  `json:"channelAccountName"`   //渠道账户姓名
	OrderAmount             float64 `json:"orderAmount"`          // 订单金额
	ActualAmount            float64 `json:"actualAmount"`         // 实际金额
	TransferHandlingFee     float64 `json:"transferHandlingFee"`
	MerchantOrderNo         string  `json:"merchantOrderNo"` //商戶訂單編號
	ChannelOrderNo          string  `json:"channelOrderNo"`  //渠道订单编号
	Source                  string  `json:"source"`          //1:平台 2:API
	SourceOrderNo           string  `json:"sourceOrderNo"`   //來源訂單編號(From NC)
	CallBackStatus          string  `json:"callBackStatus, optional"`
	NotifyUrl               string  `json:"notifyUrl, optional"`
	PageUrl                 string  `json:"pageUrl, optional"`
	PersonProcessStatus     string  `json:"personProcessStatus, optional"` //人工处理状态：(0:待處理1:處理中2:成功3:失敗 10:不需处理)
	IsMerchantCallback      string  `json:"isMerchantCallback, optional"`  //是否已经回调商户(0：否、1:是、2:不需回调)(透过API需提供的资讯)
	CreatedBy               string  `json:"createdBy, optional"`
	UpdatedBy               string  `json:"updatedBy, optional"`
}

type InternalChargeOrder struct {
	ID                      int64  `json:"id"`
	MerchantCode            string `json:"merchantCode"`
	OrderNo                 string `json:"orderNo"`
	MerchantAccountName     string `json:"merchantAccountName"`
	ChannelAccountName      string `json:"channelAccountName"`
	CurrencyCode            string `json:"currencyCode"`
	OrderAmount             string `json:"OrderAmount"`
	CreatedAt               string `json:"createdAt"`
	InternalChargeOrderPath string `json:"internalChargeOrderPath"`
	MerchantBankAccount     string `json:"merchantBankAccount"`
	ChannelBankAccount      string `json:"channelBankAccount"`
}

type WithDrawOrder struct {
	ID                   int64   `json:"id"`
	MerchantCode         string  `json:"merchantCode"`
	OrderNo              string  `json:"orderNo"`
	MerchantAccountName  string  `json:"accountName"`
	MerchantBankAccount  string  `json:"merchantBankAccount"`
	MerchantBankProvince string  `json:"merchantBankProvince"`
	MerchantBankCity     string  `json:"merchantBankCity"`
	CurrencyCode         string  `json:"currencyCode"`
	OrderAmount          string  `json:"OrderAmount"`
	CreatedAt            string  `json:"createdAt"`
	HandlingFee          float64 `json:"handlingfee"`
	Status               string  `json:"status"`
}

type BatchCheck struct {
	CurrencyCode         string  `json:"currencyCode"`
	MerchantAccountName  string  `json:"merchantAccountName"`
	MerchantBankAccount  string  `json:"merchantBankAccount"`
	MerchantBankNo       string  `json:"merchantBankNo"`
	MerchantBankName     string  `json:"merchantBankName"`
	MerchantBankProvince string  `json:"merchantBankProvince"`
	MerchantBankCity     string  `json:"merchantBankCity"`
	TransferAmount       float64 `json:"transferAmount, optional"`
	OrderAmount          float64 `json:"orderAmount, optional"`
	Valid                bool    `json:"valid, optional"`
	MinWithdrawCharge    float64 `json:"minWithdrawCharge, optional"`
	MaxWithdrawCharge    float64 `json:"maxWithdrawCharge, optional"`
}

type UploadImageRequest struct {
}

type ChannelWithdraw struct {
	ChannelCode    string  `json:"channelCode"`
	WithdrawAmount float64 `json:"withdrawAmount"`
}

type OrderChannels struct {
	ID                  int64       `json:"id"`
	OrderNo             string      `json:"orderNo"`
	ChannelCode         string      `json:"channelCode"`
	ChannelData         ChannelData `json:"channelData, optional" gorm:"foreignKey:Code;references:ChannelCode"`
	OrderAmount         float64     `json:"orderAmount"`
	TransferHandlingFee float64     `json:"transferHandlingFee"`
	HandlingFee         float64     `json:"handlingFee"`
	Fee                 float64     `json:"fee,optional"`
	Status              string      `json:"status,optional"`
	ErrorMsg            string      `json:"error_msg,optional"`
	ChannelOrderNo      string      `json:"channel_order_no,optional"`
}

type OrderInternalCreateRequest struct {
	Type                 string  `json:"type, optional"`
	MerchantBankAccount  string  `json:"merchantBankAccount"`
	MerchantAccountName  string  `json:"merchantAccountName"`
	MerchantBankNo       string  `json:"merchantBankNo"`
	MerchantBankName     string  `json:"merchantBankName, optional"`
	MerchantBankProvince string  `json:"merchantBankProvince, optional"`
	MerchantBankCity     string  `json:"merchantBankCity, optional"`
	ChannelBankAccount   string  `json:"channelBankAccount, optional"`
	ChannelAccountName   string  `json:"channelAccountName, optional"`
	ChannelBankNo        string  `json:"channelBankNo, optional"`
	CurrencyCode         string  `json:"currencyCode, optional"`
	OrderAmount          float64 `json:"orderAmount"`
}

type OrderProxyCreateRequeset struct {
	Index                string  `json:"index"`
	MerchantBankAccount  string  `json:"merchantBankAccount"`
	MerchantBankNo       string  `json:"merchantBankNo"`
	MerchantBankName     string  `json:"merchantBankName, optional"`
	MerchantAccountName  string  `json:"merchantAccountName"`
	MerchantBankProvince string  `json:"merchantBankProvince, optional"`
	MerchantBankCity     string  `json:"merchantBankCity, optional"`
	CurrencyCode         string  `json:"currencyCode"`
	OrderAmount          float64 `json:"orderAmount"`
}

type OrderWithdrawCreateRequest struct {
	MerchantAccountName  string  `json:"merchantAccountName"`
	MerchantBankAccount  string  `json:"merchantBankAccount"`
	MerchantBankNo       string  `json:"merchantBankNo"`
	MerchantBankName     string  `json:"merchantBankName, optional"`
	MerchantBankProvince string  `json:"merchantBankProvince, optional"`
	MerchantBankCity     string  `json:"merchantBankCity, optional"`
	CurrencyCode         string  `json:"currencyCode, optional"`
	OrderAmount          float64 `json:"orderAmount"`
	Index                string  `json:"index, optional"`
}

type MultipleOrderProxyCreateRequest struct {
	List []OrderProxyCreateRequeset `json:"list"`
}

type MultipleOrderWithdrawCreateRequest struct {
	List []OrderWithdrawCreateRequest `json:"list"`
}

type MultipleOrderCreateResponse struct {
	Index []string `json:"index"`
	Errs  []string `json:"errs"`
}

type OrderUpdateRequest struct {
	ID      int64  `json:"id"`
	Status  string `json:"status"`
	OrderNo string `json:"orderNo"`
	Memo    string `json:"memo, optional"`
}

type WithdrawOrderUpdateRequest struct {
	List []ChannelWithdraw `json:"list"`
	OrderUpdateRequest
}

type InternalChargeOrderQueryAllRequest struct {
	MerchantCode string `json:"merchantCode"`
	Status       string `json:"status"`
	StartAt      string `json:"startAt"`
	EndAt        string `json:"endAt"`
	PageNum      int    `json:"pageNum" gorm:"-"`
	PageSize     int    `json:"pageSize" gorm:"-"`
}

type InternalChargeOrderQueryAllResponse struct {
	List     []InternalChargeOrder `json:"list"`
	PageNum  int                   `json:"pageNum" gorm:"-"`
	PageSize int                   `json:"pageSize" gorm:"-"`
	RowCount int64                 `json:"rowCount"`
}

type WithDrawOrderQueryAllRequest struct {
	MerchantCode string `json:"merchantCode"`
	Status       string `json:"status"`
	StartAt      string `json:"startAt"`
	EndAt        string `json:"endAt"`
	PageNum      int    `json:"pageNum" gorm:"-"`
	PageSize     int    `json:"pageSize" gorm:"-"`
}

type WithdrawOrderQueryAllResponse struct {
	List     []WithDrawOrder `json:"list"`
	PageNum  int             `json:"pageNum" gorm:"-"`
	PageSize int             `json:"pageSize" gorm:"-"`
	RowCount int64           `json:"rowCount"`
}

type BatchCheckOrderRequest struct {
	List []BatchCheck `json:"list"`
	Type string       `json:"type"`
}

type BatchCheckOrderRespsonse struct {
	List        []BatchCheck `json:"list"`
	TotalPrice  float64      `json:"totalPrice"`
	HandlingFee float64      `json:"handlingFee"`
}

type OrderQueryWaitAuditNumberRequest struct {
	Type    string `json:"type"`
	StartAt string `json:"startAt"`
	EndAt   string `json:"endAt"`
}

type OrderQueryWaitAuditNumberResponse struct {
	WaitAuditNumber int64 `json:"waitAuditNumber"`
}

type OrderQueryMerchantChannelFeeRequest struct {
	CurrencyCode string `json:"currencyCode"`
	Type         string `json:"type"`
}

type OrderQueryMerchantCurrencyAndBanksResponse struct {
}

type WithdrawOrderChannelRequest struct {
	CurrencyCode string `json:"currencyCode"`
}

type WithdrawOrderChannelResponse struct {
	List []ChannelCodeAndHandlingFee `json:"list"`
}

type WithdrawOrderFeeRequest struct {
	CurrencyCode string `json:"currencyCode"`
	UserAccount  string `json:"userAccount"`
}

type WithdrawOrderFeeResponse struct {
	HandlerFee        float64 `json:"handlerFee"`
	MinWithdrawCharge float64 `json:"minWithdrawCharge"`
	MaxWithdrawCharge float64 `json:"maxWithdrawCharge"`
	Balance           float64 `json:"balance"`
}

type WithdrawOrderUpdateReviewProcssRequest struct {
	OrderNo string `json:"orderNo"`
}

type WithdrawVerifyWayResponse struct {
	PayingValidatedType string `json:"payingValidatedType"`
}

type WithdrawVerifyPasswordRequest struct {
	WithdrawPassword string `json:"withdrawPassword"`
}

type WithdrawVerifyPasswordResponse struct {
}

type OrderAction struct {
	ID          int64  `json:"id"`
	OrderNo     string `json:"orderNo"`
	Action      string `json:"action"`
	UserAccount string `json:"userAccount"`
	Comment     string `json:"comment"`
	CreatedAt   string `json:"createdAt"`
}

type OrderActionQueryAllRequest struct {
	OrderNo     string `json:"orderNo, optional"`
	UserAccount string `json:"userAccount, optional"`
	StartAt     string `json:"startAt, optional"`
	EndAt       string `json:"endAt, optional"`
	PageNum     int    `json:"pageNum, optional" gorm:"-"`
	PageSize    int    `json:"pageSize, optional" gorm:"-"`
}

type OrderActionQueryAllResponse struct {
	List     []OrderAction `json:"list"`
	PageNum  int           `json:"pageNum" gorm:"-"`
	PageSize int           `json:"pageSize" gorm:"-"`
	RowCount int64         `json:"rowCount"`
}

type ReceiptRecord struct {
	ID                   int64       `json:"id"`
	MerchantCode         string      `json:"merchantCode"`
	OrderNo              string      `json:"orderNo"`
	MerchantOrderNo      string      `json:"merchantOrderNo"`
	MerchantBankName     string      `json:"merchantBankName"`
	MerchantBankProvince string      `json:"merchantBankProvince"`
	MerchantAccountName  string      `json:"merchantAccountName"`
	MerchantBankAccount  string      `json:"merchantBankAccount"`
	CurrencyCode         string      `json:"currencyCode"`
	ChannelPayTypesCode  string      `json:"channelPayTypesCode"`
	ChannelCode          string      `json:"channelCode"`
	Type                 string      `json:"type"`
	PayTypeCode          string      `json:"payTypeCode"`
	BalanceType          string      `json:"balanceType"`
	OrderAmount          float64     `json:"orderAmount"`
	ActualAmount         float64     `json:"actualAmount"` // 实际金额
	TransferAmount       float64     `json:"transferAmount"`
	TransferHandlingFee  float64     `json:"transferHandlingFee"`
	HandlingFee          float64     `json:"handlingFee"`
	Fee                  float64     `json:"fee"`
	BeforeBalance        float64     `json:"beforeBalance"`
	Balance              float64     `json:"balance"`
	Status               string      `json:"status"`
	CallBackStatus       string      `json:"callBackStatus"`
	IsMerchantCallback   string      `json:"isMerchantCallback"`
	Memo                 string      `json:"memo"`
	PayTypeData          PayType     `json:"payTypeData, optional" gorm:"foreignKey:Code;references:PayTypeCode"`
	ChannelData          ChannelData `json:"channelData, optional" gorm:"foreignKey:Code;references:ChannelCode"`
	TransAt              string      `json:"transAt"`
	CreatedAt            string      `json:"createdAt"`
}

type DeductRecord struct {
	ID                  int64           `json:"id"`
	MerchantCode        string          `json:"merchantCode"`
	OrderNo             string          `json:"orderNo"`
	MerchantOrderNo     string          `json:"merchantOrderNo"`
	CurrencyCode        string          `json:"currencyCode"`
	Type                string          `json:"type"`
	OrderChannels       []OrderChannels `json:"orderChannels, optional" gorm:"foreignKey:OrderNo;references:OrderNo"`
	OrderAmount         float64         `json:"orderAmount"`
	TransferAmount      float64         `json:"transferAmount"`
	TransferHandlingFee float64         `json:"transferHandlingFee"`
	HandlingFee         float64         `json:"handlingFee"`
	Fee                 float64         `json:"fee"`
	BeforeBalance       float64         `json:"beforeBalance"`
	Balance             float64         `json:"balance"`
	Status              string          `json:"status"`
	Memo                string          `json:"memo"`
	CreatedAt           string          `json:"createdAt"`
}

type PersonalRepayment struct {
	ID                    int64  `json:"id"`
	OrderNo               string `json:"orderNo"`
	MerchantOrderNo       string `json:"merchantOrderNo"`
	ChannelCode           string `json:"channelCode"`
	MerchantBankAccount   string `json:"merchantBankAccount"`
	MerchantAccountName   string `json:"merchantAccountName"`
	MerchantBankNo        string `json:"merchantBankNo"`
	MerchantBankProvince  string `json:"merchantBankProvince"`
	MerchantBankCity      string `json:"merchantBankCity"`
	Source                string `json:"source"`
	CreatedAt             string `json:"createdAt"`
	TransAt               string `json:"transAt"`
	PersonalProcessStatus string `json:"personalProcessStatus"`
}

type FrozenRecord struct {
	ID                  int64       `json:"id"`
	MerchantCode        string      `json:"merchantCode"`
	OrderNo             string      `json:"orderNo"`
	MerchantOrderNo     string      `json:"merchantOrderNo"`
	CurrencyCode        string      `json:"currencyCode"`
	ChannelPayTypesCode string      `json:"channelPayTypesCode"`
	ChannelCode         string      `json:"channelCode"`
	Type                string      `json:"type"`
	PayTypeCode         string      `json:"payTypeCode"`
	FrozenAmount        float64     `json:"frozenAmount"`
	Status              string      `json:"status"`
	Memo                string      `json:"memo"`
	PayTypeData         PayType     `json:"payTypeData, optional" gorm:"foreignKey:Code;references:PayTypeCode"`
	ChannelData         ChannelData `json:"channelData, optional" gorm:"foreignKey:Code;references:ChannelCode"`
	TransAt             string      `json:"transAt"`
	CreatedAt           string      `json:"createdAt"`
	UpdatedAt           string      `json:"updatedAt"`
	UpdatedBy           string      `json:"updatedBy"`
}

type ReversalRecordRequest struct {
	OrderNo string `json:"orderNo"`
	Comment string `json:"comment, optional"`
}

type ReceiptRecordQueryAllRequest struct {
	MerchantCode    string `json:"merchantCode, optional"`
	OrderNo         string `json:"orderNo, optional"`
	MerchantOrderNo string `json:"merchantOrderNo, optional"`
	DateType        string `json:"dateType, optional"`
	StartAt         string `json:"startAt, optional"`
	EndAt           string `json:"endAt, optional"`
	Status          string `json:"status, optional"`
	Type            string `json:"type, optional"`
	PayTypeCode     string `json:"payTypeCode, optional"`
	CurrencyCode    string `json:"currencyCode, optional"`
	PageNum         int    `json:"pageNum, optional" gorm:"-"`
	PageSize        int    `json:"pageSize, optional" gorm:"-"`
}

type ReceiptRecordQueryAllResponse struct {
	List     []ReceiptRecord `json:"list"`
	PageNum  int             `json:"pageNum" gorm:"-"`
	PageSize int             `json:"pageSize" gorm:"-"`
	RowCount int64           `json:"rowCount"`
}

type FrozenReceiptOrderRequest struct {
	OrderNo      string  `json:"orderNo"`
	FrozenAmount float64 `json:"frozenAmount"`
	Comment      string  `json:"comment, optional"`
}

type UnfrozenOrderRequest struct {
	OrderNo string `json:"orderNo"`
}

type MakeUpReceiptOrderRequest struct {
	OrderNo        string  `json:"orderNo"`
	ChannelOrderNo string  `json:"channelOrderNo"`
	Amount         float64 `json:"amount"`
}

type DeductRecordQueryAllRequest struct {
	MerchantCode    string `json:"merchantCode, optional"`
	OrderNo         string `json:"orderNo, optional"`
	MerchantOrderNo string `json:"merchantOrderNo, optional"`
	DateType        string `json:"dateType, optional"`
	StartAt         string `json:"startAt, optional"`
	EndAt           string `json:"endAt, optional"`
	Status          string `json:"status, optional"`
	Type            string `json:"type, optional"`
	ChannelCode     string `json:"channelCode, optional"`
	CurrencyCode    string `json:"currencyCode, optional"`
	Source          string `json:"source, optional"`
	PageNum         int    `json:"pageNum, optional" gorm:"-"`
	PageSize        int    `json:"pageSize, optional" gorm:"-"`
}

type DeductRecordQueryAllResponse struct {
	List     []DeductRecord `json:"list"`
	PageNum  int            `json:"pageNum" gorm:"-"`
	PageSize int            `json:"pageSize" gorm:"-"`
	RowCount int64          `json:"rowCount"`
}

type PersonalRepaymentResponse struct {
	List     []PersonalRepayment `json:"list"`
	PageNum  int                 `json:"pageNum" gorm:"-"`
	PageSize int                 `json:"pageSize" gorm:"-"`
	RowCount int64               `json:"rowCount"`
}

type FrozenRecordQueryAllResponse struct {
	List     []FrozenRecord `json:"list"`
	PageNum  int            `json:"pageNum" gorm:"-"`
	PageSize int            `json:"pageSize" gorm:"-"`
	RowCount int64          `json:"rowCount"`
}

type OrderChannelRecordReqeust struct {
	OrderNo string `json:"orderNo"`
}

type OrderChannelRecordResponse struct {
	List []OrderChannels `json:"list"`
}

type OrderFeeProfitQueryAllResponse struct {
	List     []OrderFeeProfit `json:"list"`
	PageNum  int              `json:"pageNum" gorm:"-"`
	PageSize int              `json:"pageSize" gorm:"-"`
	RowCount int64            `json:"rowCount"`
}

type ChannelDataCreateRequest struct {
	Code                    string           `json:"code, optional"`
	Name                    string           `json:"name" valiate: "required"`
	ProjectName             string           `json:"projectName, optional"`
	IsProxy                 string           `json:"isProxy" valiate: "required"`
	IsNZPre                 string           `json:"isNzPre" valiate: "required"`
	ApiUrl                  string           `json:"apiUrl, optional"`
	CurrencyCode            string           `json:"currencyCode" valiate: "required"`
	ChannelWithdrawCharge   float64          `json:"channelWithdrawCharge, optional"`
	Balance                 float64          `json:"balance, optional"`
	Status                  string           `json:"status, optional"`
	Device                  string           `json:"device,optional"`
	CreatedAt               string           `json:"createdAt, optional"`
	UpdatedAt               string           `json:"updatedAt, optional"`
	MerId                   string           `json:"merId, optional"`
	MerKey                  string           `json:"merKey , optional"`
	PayUrl                  string           `json:"payUrl, optional"`
	PayQueryUrl             string           `json:"payQueryUrl, optional"`
	PayQueryBalanceUrl      string           `json:"payQueryBalanceUrl, optional"`
	ProxyPayUrl             string           `json:"proxyPayUrl, optional"`
	ProxyPayQueryUrl        string           `json:"proxyPayQueryUrl, optional"`
	ProxyPayQueryBalanceUrl string           `json:"proxyPayQueryBalanceUrl, optional"`
	WhiteList               string           `json:"whiteList, optional"`
	PayTypeMapList          []PayTypeMap     `json:"payTypeMapList, optional" gorm:"-"`
	PayTypeMap              string           `json:"payTypeMap, optional"`
	ChannelPayTypeList      []ChannelPayType `json:"channelPayTypeList, optional" gorm:"foreignKey:ChannelCode;references:Code"`
	ChannelPort             string           `json:"channelPort, optional"`
	WithdrawBalance         float64          `json:"withdrawBalance, optional"`
	ProxypayBalance         float64          `json:"proxypayBalance, optional"`
	BankCodeMapList         []BankCodeMap    `json:"bankCodeMapList, optional" gorm:"-"`
	Banks                   []Bank           `json:"banks, optional" gorm:"many2many:ch_channel_banks;foreignKey:Code;joinForeignKey:channel_code;references:bank_no;joinReferences:bank_no"`
}

type ChannelDataUpdateRequest struct {
	ID                      int64            `json:"id, optional"`
	Code                    string           `json:"code" valiate: "required"`
	Name                    string           `json:"name" valiate: "required"`
	ProjectName             string           `json:"projectName" valiate: "required"`
	IsProxy                 string           `json:"isProxy" valiate: "required"`
	IsNZPre                 string           `json:"isNzPre" valiate: "required"`
	ApiUrl                  string           `json:"apiUrl, optional"`
	CurrencyCode            string           `json:"currencyCode, optional"`
	ChannelWithdrawCharge   float64          `json:"channelWithdrawCharge, optional"`
	Balance                 float64          `json:"balance, optional"`
	Status                  string           `json:"status, optional"`
	CreatedAt               string           `json:"createdAt, optional"`
	UpdatedAt               string           `json:"updatedAt, optional"`
	Device                  string           `json:"device,optional"`
	MerId                   string           `json:"merId, optional"`
	MerKey                  string           `json:"merKey , optional"`
	PayUrl                  string           `json:"payUrl, optional"`
	PayQueryUrl             string           `json:"payQueryUrl, optional"`
	PayQueryBalanceUrl      string           `json:"payQueryBalanceUrl, optional"`
	ProxyPayUrl             string           `json:"proxyPayUrl, optional"`
	ProxyPayQueryUrl        string           `json:"proxyPayQueryUrl, optional"`
	ProxyPayQueryBalanceUrl string           `json:"proxyPayQueryBalanceUrl, optional"`
	WhiteList               string           `json:"whiteList, optional"`
	PayTypeMapList          []PayTypeMap     `json:"payTypeMapList, optional" gorm:"-"`
	PayTypeMap              string           `json:"payTypeMap, optional"`
	ChannelPayTypeList      []ChannelPayType `json:"channelPayTypeList, optional" gorm:"foreignKey:ChannelCode;references:Code"`
	ChannelPort             string           `json:"channelPort, optional"`
	WithdrawBalance         float64          `json:"withdrawBalance, optional"`
	ProxypayBalance         float64          `json:"proxypayBalance, optional"`
	BankCodeMapList         []BankCodeMap    `json:"bankCodeMapList, optional" gorm:"-"`
	Banks                   []Bank           `json:"banks, optional" gorm:"many2many:ch_channel_banks;foreignKey:Code;joinForeignKey:channel_code;references:bank_no;joinReferences:bank_no"`
}

type ChannelDataDeleteRequest struct {
	ID   int64  `json:"id, optional"`
	Code string `json:"code",valiate: "required"`
}

type ChannelDataQueryRequest struct {
	ID   int64  `json:"id, optional"`
	Code string `json:"code",valiate: "required"`
}

type ChannelDataQueryResponse struct {
	ChannelData
}

type ChannelDataQueryAllRequest struct {
	Code         string `json:"code,optional"`
	Name         string `json:"name,optional"`
	CurrencyCode string `json:"currencyCode,optional"`
	Status       string `json:"status, optional"`
	IsProxy      string `json:"isProxy, optional"`
	Device       string `json:"device,optional"`
	PageNum      int    `json:"pageNum" gorm:"-"` //gorm:"-" 忽略這參數檢查是否有給
	PageSize     int    `json:"pageSize" gorm:"-"`
}

type ChannelDataQueryAllResponse struct {
	List     []ChannelData `json:"list"`
	PageNum  int           `json:"pageNum"`
	PageSize int           `json:"pageSize"`
	RowCount int64         `json:"rowCount"`
}

type ChannelDataGetDropDownListRequest struct {
	Code         string   `json:"code,optional"`
	Name         string   `json:"name,optional"`
	Status       []string `json:"status,optional"`
	CurrencyCode string   `json:"currencyCode,optional"`
}

type ChannelDataDropDownItem struct {
	Id   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type ChannelCodeAndHandlingFee struct {
	Code                  string `json:"channelCode"`
	ChannelWithdrawCharge string `json:"channelWithdrawCharge"`
	Name                  string `json:"name"`
}

type ChannelPayTypeCreateRequest struct {
	Code              string  `json:"code" valiate: "required"`
	ChannelCode       string  `json:"channelCode" valiate: "required"`
	PayTypeCode       string  `json:"payTypeCode" valiate: "required"`
	Fee               float64 `json:"fee, optional"`
	HandlingFee       float64 `json:"handlingFee, optional"`
	MaxInternalCharge float64 `json:"maxInternalCharge, optional"`
	DailyTxLimit      float64 `json:"dailyTxLimit, optional"`
	SingleMinCharge   float64 `json:"singleMinCharge, optional"`
	SingleMaxCharge   float64 `json:"singleMaxCharge, optional"`
	FixedAmount       string  `json:"fixedAmount, optional"`
	BillDate          int64   `json:"billDate, optional"`
	Status            string  `json:"status" valiate: "required"`
	IsProxy           string  `json:"isProxy, optional"`
	Device            string  `json:"device, optional"`
	MapCode           string  `json:"mapCode, optional"`
}

type ChannelPayTypeUpdateRequest struct {
	ID                int64   `json:"id, optional"`
	Code              string  `json:"code" valiate: "required"`
	ChannelCode       string  `json:"channelCode" valiate: "required"`
	PayTypeCode       string  `json:"payTypeCode" valiate: "required"`
	Fee               float64 `json:"fee, optional"`
	HandlingFee       float64 `json:"handlingFee, optional"`
	MaxInternalCharge float64 `json:"maxInternalCharge, optional"`
	DailyTxLimit      float64 `json:"dailyTxLimit, optional"`
	SingleMinCharge   float64 `json:"singleMinCharge, optional"`
	SingleMaxCharge   float64 `json:"singleMaxCharge, optional"`
	FixedAmount       string  `json:"fixedAmount, optional"`
	BillDate          int64   `json:"billDate, optional"`
	Status            string  `json:"status" valiate: "required"`
	IsProxy           string  `json:"isProxy, optional"`
	Device            string  `json:"device, optional"`
	MapCode           string  `json:"mapCode, optional"`
}

type ChannelPayTypeQueryAllRequest struct {
	ChannelName string `json:"channelName,optional"`
	PayTypeName string `json:"payTypeName,optional"`
	Status      string `json:"status,optional"`
	IsProxy     string `json:"isProxy,optional"`
	Currency    string `json:"currency,optional"`
	PageNum     int    `json:"pageNum" gorm:"-"` //gorm:"-" 忽略這參數檢查是否有給
	PageSize    int    `json:"pageSize" gorm:"-"`
}

type ChannelPayTypeQueryAllResponse struct {
	List     []ChannelPayTypeQueryResponse `json:"list"`
	PageNum  int                           `json:"pageNum"`
	PageSize int                           `json:"pageSize"`
	RowCount int64                         `json:"rowCount"`
}

type ChannelPayTypeQueryRequest struct {
	ID int64 `json:"id"`
}

type ChannelPayTypeQueryResponse struct {
	ChannelPayType
	CurrencyCode string `json:"currencyCode"`
	PayTypeName  string `json:"payTypeName"`
	ChannelName  string `json:"channelName"`
}

type ChannelPayTypeDeleteRequest struct {
	ID int64 `json:"id"`
}

type PayTypeCreateRequest struct {
	Code         string        `json:"code, optional" valiate: "required"`
	Name         string        `json:"name, optional" valiate: "required"`
	Currency     string        `json:"currency, optional"`
	ImgUrl       string        `json:"imgUrl, optional"`
	ChannelDatas []ChannelData `json:"channelDatas, optional" gorm:"many2many:ch_channel_pay_types:foreignKey:code;joinForeignKey:pay_type_code;references:code;joinReferences:channel_code"`
}

type PayTypeUpdateRequest struct {
	ID           int64         `json:"id" valiate: "required"`
	Code         string        `json:"code, optional" valiate: "required"`
	Name         string        `json:"name, optional" valiate: "required"`
	Currency     string        `json:"currency, optional"`
	ImgUrl       string        `json:"imgUrl, optional"`
	ChannelDatas []ChannelData `json:"channelDatas, optional" gorm:"many2many:ch_channel_pay_types:foreignKey:code;joinForeignKey:pay_type_code;references:code;joinReferences:channel_code"`
}

type PayTypeDeleteRequest struct {
	ID int64 `json:"id"`
}

type PayTypeQueryRequest struct {
	ID int64 `json:"id"`
}

type PayTypeQueryResponse struct {
	PayType
}

type PayTypeQueryAllRequest struct {
	Code     string `json:"code,optional"`
	Name     string `json:"name,optional"`
	Currency string `json:"currency,optional"`
	PageNum  int    `json:"pageNum, optional" gorm:"-"` //gorm:"-" 忽略這參數檢查是否有給
	PageSize int    `json:"pageSize, optional" gorm:"-"`
}

type PayTypeQueryAllResponse struct {
	List     []PayType `json:"list"`
	PageNum  int       `json:"pageNum"`
	PageSize int       `json:"pageSize"`
	RowCount int64     `json:"rowCount"`
}

type ChannelBank struct {
	ID          int64  `json:"id, optional"`
	ChannelCode string `json:"channelCode, optional"`
	BankNo      string `json:"bankNo, optional"`
	MapCode     string `json:"mapCode, optional"`
}

type ChannelBankCreateRequest struct {
	ChannelCode string `json:"channelCode, optional"`
	BankNo      string `json:"bankNo" valiate: "required"`
	MapCode     string `json:"mapCode"`
}

type ChannelBankUpdateRequest struct {
	ID          int64  `json:"id, optional"`
	ChannelCode string `json:"channelCode" valiate: "required"`
	BankNo      string `json:"bankNo" valiate: "required"`
	MapCode     string `json:"mapCode"`
}

type ChannelBankDeleteRequest struct {
	ID int64 `json:"id"`
}

type ChannelBankQueryRequest struct {
	ID          int64  `json:"id"`
	ChannelCode string `json:"channelCode"`
	BankNo      string `json:"bankNo"`
}

type ChannelBankQueryResponse struct {
	ChannelBank ChannelBank `json:"channelBank"`
}

type ChannelBankQueryAllRequest struct {
	ID          int64  `json:"id"`
	ChannelCode string `json:"channelCode"`
	BankNo      string `json:"bankNo"`
}

type ChannelBankQueryAllResponse struct {
	ChannelBankList []ChannelBank `json:"channelBankList"`
}

type TxLog struct {
	ID              int64  `json:"id"`
	MerchantCode    string `json:"merchantCode, optional"`
	OrderNo         string `json:"orderNo, optional"`
	MerchantOrderNo string `json:"merchantOrderNo, optional"`
	ChannelOrderNo  string `json:"channelOrderNo, optional"`
	LogType         string `json:"logType, optional"`
	LogSource       string `json:"logSource, optional"`
	Content         string `json:"content, optional"`
	Log             string `json:"log, optional"`
	CreatedAt       string `json:"createdAt, optional"`
	ErrorCode       string `json:"errorCode, optional"`
	ErrorMsg        string `json:"errorMsg, optional"`
	TraceId         string `json:"traceId, optional"`
}

type TransactionLogData struct {
	MerchantNo      string      `json:"merchantNo"`
	MerchantOrderNo string      `json:"merchantOrderNo"`
	OrderNo         string      `json:"orderNo"`
	LogType         string      `json:"logType"`
	LogSource       string      `json:"logSource"`
	Content         interface{} `json:"content"`
	ErrCode         string      `json:"errCode"`
	ErrMsg          string      `json:"errMsg"`
	TraceId         string      `json:"traceId, optional"`
}

type MerchantQueryOrderStatementRequestX struct {
	MerchantQueryOrderStatementRequest
	Ip string `json:"ip, optional"`
}
