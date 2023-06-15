package model

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/types"
	"gorm.io/gorm"
)

type merchantBalance struct {
	MyDB  *gorm.DB
	Table string
}

func NewMerchantBalance(mydb *gorm.DB, t ...string) *merchantBalance {
	table := "mc_merchant_balances"
	if len(t) > 0 {
		table = t[0]
	}
	return &merchantBalance{
		MyDB:  mydb,
		Table: table,
	}
}

func (u *merchantBalance) IsExistFromMerchantBalance(merchantCode, currencyCode string) (isExist bool, err error) {
	err = u.MyDB.Table(u.Table).
		Select("count(*) > 0").
		Where("merchant_code = ? AND currency_code = ?", merchantCode, currencyCode).
		Find(&isExist).Error
	return
}

func (u *merchantBalance) CreateMerchantBalances(merchantCode, currencyCode string) (err error) {
	var isExist bool

	if isExist, err = u.IsExistFromMerchantBalance(merchantCode, currencyCode); err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	} else if isExist {
		return
	}

	u.createMerchantBalance(merchantCode, currencyCode, "DFB")
	u.createMerchantBalance(merchantCode, currencyCode, "XFB")
	u.createMerchantBalance(merchantCode, currencyCode, "YJB")

	return
}

func (u *merchantBalance) createMerchantBalance(merchantCode, currencyCode, balanceType string) (err error) {

	var merchantBalance = types.MerchantBalance{
		MerchantCode: merchantCode,
		CurrencyCode: currencyCode,
		BalanceType:  balanceType,
	}

	if err = u.MyDB.Table(u.Table).Create(&types.MerchantBalanceX{
		MerchantBalance: merchantBalance,
	}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
