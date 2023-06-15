package model

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/types"
	"gorm.io/gorm"
)

type merchantCurrency struct {
	MyDB  *gorm.DB
	Table string
}

func NewMerchantCurrency(mydb *gorm.DB, t ...string) *merchantCurrency {
	table := "mc_merchant_currencies"
	if len(t) > 0 {
		table = t[0]
	}
	return &merchantCurrency{
		MyDB:  mydb,
		Table: table,
	}
}

func (c *merchantCurrency) GetByMerchantCode(code string, status string) (merchantCurrencies []types.MerchantCurrency, err error) {

	db := c.MyDB
	if len(status) > 0 {
		db = db.Where("status = ?", status)
	}
	db = db.Where("merchant_code = ?", code)

	err = db.Table(c.Table).
		Order("currency_code").
		Find(&merchantCurrencies).Error

	return
}

func (c *merchantCurrency) CreateMerchantCurrency(merchantCode, currencyCode, status string) (err error) {
	if err = c.MyDB.Table(c.Table).Create(&types.MerchantCurrencyCreate{
		MerchantCurrencyCreateRequest: types.MerchantCurrencyCreateRequest{
			MerchantCode: merchantCode,
			CurrencyCode: currencyCode,
			Status:       status,
		},
	}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func (c *merchantCurrency) IsExistFromMerchantCurrency(merchantCode, currencyCode string) (isExist bool, err error) {
	err = c.MyDB.Table(c.Table).
		Select("count(*) > 0").
		Where("merchant_code = ? AND currency_code = ?", merchantCode, currencyCode).
		Find(&isExist).Error
	return
}
