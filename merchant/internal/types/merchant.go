package types

import (
	"com.copo/bo_service/common/gormx"
	"time"
)

type MerchantQueryChannelRequestX struct {
	MerchantQueryChannelRequest
	MyIp string `json:"my_ip, optional"`
}

type MerchantX struct {
	Merchant
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MerchantCreate struct {
	MerchantCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MerchantQueryListViewRequestX struct {
	MerchantQueryListViewRequest
	Language string        `json:"language, optional" gorm:"-"`
	Orders   []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type MerchantConfigureRateListRequestX struct {
	MerchantConfigureRateListRequest
	Language string        `json:"language, optional" gorm:"-"`
	Orders   []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type MerchantQueryRateListViewRequestX struct {
	MerchantQueryRateListViewRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type MerchantQueryAllRequestX struct {
	MerchantQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type MerchantCurrencyQueryAllRequestX struct {
	MerchantCurrencyQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type MerchantBalanceRecordQueryAllRequestX struct {
	MerchantBalanceRecordQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type MerchantCurrencyCreate struct {
	MerchantCurrencyCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MerchantCurrencyUpdate struct {
	MerchantCurrencyUpdateRequest
	CreatedAt time.Time `json:"createdAt, optional"`
	UpdatedAt time.Time `json:"updatedAt, optional"`
}

type MerchantUpdateCurrenciesRequestX struct {
	MerchantUpdateCurrenciesRequest
	Currencies []MerchantCurrencyUpdate `json:"currencies"`
}

type MerchantUpdate struct {
	MerchantUpdateRequest
	CreatedAt time.Time `json:"createdAt, optional"`
	UpdatedAt time.Time `json:"updatedAt, optional"`
}

type MerchantUpdate2 struct {
	Merchant
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MerchantBalanceCreate struct {
	MerchantBalanceCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MerchantBalanceUpdate struct {
	MerchantBalanceUpdateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MerchantChannelRateConfigure struct {
	MerchantChannelRateConfigureRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MerchantBindBankCreate struct {
	MerchantBindBankCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MerchantBindBankUpdate struct {
	MerchantBindBankUpdateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChannelChangeNotify struct {
	ID                    int64  `json:"id"`
	MerchantCode          string `json:"merchantCode"`
	IsChannelChangeNotify string `json:"isChannelChangeNotify"`
	NotifyUrl             string `json:"notifyUrl"`
	LastNotifyMessage     string `json:"lastNotifyMessage"`
	Status                string `json:"status"`
	CreatedBy             string `json:"createdBy"`
	UpdatedBy             string `json:"updatedBy"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type User struct {
	ID            int64      `json:"id"`
	Account       string     `json:"account"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone"`
	Avatar        string     `json:"avatar"`
	Password      string     `json:"-"`
	RegisteredAt  int64      `json:"registeredAt, optional"`
	RegisteredAtS string     `json:"registeredAtS, optional" gorm:"-"`
	LastLoginAt   int64      `json:"lastLoginAt"`
	LastLoginIP   string     `json:"lastLoginIp"`
	Status        string     `json:"status"`
	DisableDelete string     `json:"disableDelete"`
	IsLogin       string     `json:"isLogin"`
	Currencies    []Currency `json:"currencies, optional" gorm:"many2many:au_user_currencies;foreignKey:account;joinForeignKey:user_account;references:code;joinReferences:currency_code;"`
	Roles         []Role     `json:"roles" gorm:"many2many:au_user_roles;foreignKey:id;joinForeignKey:user_id;"`
	Merchants     []Merchant `json:"merchantsService" gorm:"many2many:au_user_merchants;foreignKey:account;joinForeignKey:user_account;references:code;joinReferences:merchant_code;"`
}

type UserCreateRequest struct {
	ID              int64      `json:"id, optional"`
	Account         string     `json:"account" validate:"alphanumLength=6/12"`
	Name            string     `json:"name" validate:"required,lte=32"`
	Avatar          string     `json:"avatar,optional"`
	Email           string     `json:"email, optional"`
	Phone           string     `json:"phone, optional"`
	Roles           []Role     `json:"roles, optional" gorm:"many2many:au_user_roles;foreignKey:id;joinForeignKey:user_id;"`
	Merchants       []Merchant `json:"merchants, optional" gorm:"many2many:au_user_merchants;foreignKey:account;joinForeignKey:user_account;references:code;joinReferences:merchant_code;"`
	Password        string     `json:"password" validate:"required,alphanumLength=8/16"`
	PasswordConfirm string     `json:"passwordConfirm" validate:"eqfield=Password" gorm:"-"`
	RegisteredAt    int64      `json:"registeredAt, optional"`
	Status          string     `json:"status, optional"`
	IsLogin         string     `json:"islogin, optional"`
	IsAdmin         string     `json:"isAdmin, optional"`
	DisableDelete   string     `json:"disableDelete, optional"`
	Remark          string     `json:"remark, optional"`
	Currencies      []Currency `json:"currencies, optional" gorm:"many2many:au_user_currencies;foreignKey:account;joinForeignKey:user_account;references:code;joinReferences:currency_code;"`
}

type UserUpdateRequest struct {
	ID              int64      `json:"id"`
	Account         string     `json:"account,optional" validate:"alphanumLength=6/12"`
	Name            string     `json:"name,optional" validate:"lte=32"`
	Avatar          string     `json:"avatar,optional"`
	Email           string     `json:"email, optional" validate:"email"`
	Phone           string     `json:"phone, optional"`
	OtpKey          string     `json:"otpKey, optional"`
	Qrcode          string     `json:"qrcode, optional"`
	Roles           []Role     `json:"roles, optional" gorm:"many2many:au_user_roles;foreignKey:id;joinForeignKey:user_id;"`
	Merchants       []Merchant `json:"merchantsService, optional" gorm:"many2many:au_user_merchants;foreignKey:account;joinForeignKey:user_account;references:code;joinReferences:merchant_code;"`
	OldPassword     string     `json:"oldPassword,optional" validate:"required_with=Password" gorm:"-"`
	Password        string     `json:"password,optional" validate:"required_with=OldPassword,alphanumLength=8/16"`
	PasswordConfirm string     `json:"passwordConfirm,optional" validate:"eqfield=Password" gorm:"-"`
	RegisteredAt    int64      `json:"registeredAt, optional"`
	Status          string     `json:"status, optional"`
	IsLogin         string     `json:"isLogin, optional"`
	DisableDelete   string     `json:"disableDelete, optional"`
	Currencies      []Currency `json:"currencies, optional" gorm:"many2many:au_user_currencies;foreignKey:account;joinForeignKey:user_account;references:code;joinReferences:currency_code;"`
}

type UserDeleteRequest struct {
	ID int64 `json:"id"`
}

type UserQueryRequest struct {
	ID int64 `json:"id"`
}

type UserQueryResponse struct {
	User
}

type UserQueryAllRequest struct {
	Account  string `json:"account,optional"`
	Name     string `json:"name,optional"`
	Email    string `json:"email,optional"`
	PageNum  int    `json:"pageNum" gorm:"-"`
	PageSize int    `json:"pageSize" gorm:"-"`
}

type UserQueryAllResponse struct {
	List     []User `json:"list"`
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
	RowCount int64  `json:"rowCount"`
}
