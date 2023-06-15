package merchantchannelrateservice

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/types"
	"gorm.io/gorm"
)

func GetAllMerchantChannelRateByPayType(db *gorm.DB, merchantCode, currencyCode, payTypeCode, device string) (resp []*types.MerchantOrderRateListViewX, err error) {

	db = db.Where("merchant_code = ?", merchantCode).
		Where("merchnrate_status = '1'").
		Where("designation = '1'").
		Where("chn_status = '1'").
		Where("chnpaytype_status = '1'").
		Where("currency_code = ?", currencyCode).
		Where("pay_type_code = ?", payTypeCode)

	if device == "all" || device == "All" {
		db = db.Where("device = ?", device)

	} else {
		db = db.Where("device in ?", []string{"All", device})
	}

	if err = db.Table("merchant_order_rate_list_view").Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, "数据库错误: "+err.Error())
	}

	return resp, nil
}
