package model

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/types"
	"fmt"
	"gorm.io/gorm"
	"regexp"
	"strconv"
)

type Merchant struct {
	MyDB  *gorm.DB
	Table string
}

func NewMerchant(mydb *gorm.DB, t ...string) *Merchant {
	table := "mc_merchants"
	if len(t) > 0 {
		table = t[0]
	}
	return &Merchant{
		MyDB:  mydb,
		Table: table,
	}
}

func (m *Merchant) GetMerchant(id int64) (merchant *types.Merchant, err error) {
	err = m.MyDB.Table(m.Table).
		Preload("Users").
		Preload("MerchantBalances").
		Preload("MerchantCurrencies").
		Take(&merchant, id).Error
	return
}
func (m *Merchant) GetMerchantByCode(code string) (merchant *types.Merchant, err error) {
	err = m.MyDB.Table(m.Table).
		Preload("Users").
		Preload("Users.Currencies").
		Preload("MerchantBalances").
		Preload("MerchantCurrencies").
		Where("code = ? ", code).
		Take(&merchant).Error

	return
}

func (m *Merchant) GetDescendantAgentsByCode(merchantCode string, isIncludeItself bool) (merchants []types.Merchant, err error) {
	var merchant *types.Merchant
	if merchant, err = m.GetMerchantByCode(merchantCode); err != nil {
		return
	}
	if merchant.AgentLayerCode == "" {
		return merchants, errorz.New(response.AGENT_LAYER_NO_GET_ERROR)
	}
	return m.GetDescendantAgents(merchant.AgentLayerCode, isIncludeItself)
}

// GetDescendantAgents 取得所有子孫商戶 (可選擇是否包含自己)
func (m *Merchant) GetDescendantAgents(agentLayerCode string, isIncludeItself bool) (merchants []types.Merchant, err error) {

	db := m.MyDB.Table(m.Table)

	if !isIncludeItself {
		db = db.Where("agent_layer_code != ?", agentLayerCode)
	}

	err = db.Where("agent_layer_code LIKE ?", agentLayerCode+"%").
		Preload("Users").
		Preload("Users.Currencies").
		Order("agent_layer_code").
		Find(&merchants).Error
	return
}

// GetSubAgents 取得所有子商戶
func (m *Merchant) GetSubAgents(code string) (merchants []types.Merchant, err error) {
	err = m.MyDB.Table(m.Table).Where("agent_parent_code = ?", code).
		Order("agent_layer_code").
		Find(&merchants).Error
	return
}

func (m *Merchant) QueryMerchants(req types.MerchantQueryAllRequestX) (merchants []types.Merchant, count int64, err error) {

	db := m.MyDB

	if len(req.Code) > 0 {
		db = db.Where("code like ?", "%"+req.Code+"%")
	}
	if len(req.AgentLayerCode) > 0 {
		db = db.Where("agent_layer_code like ?", "%"+req.AgentLayerCode+"%")
	}
	if len(req.AccountName) > 0 {
		db = db.Where("account_name like ?", "%"+req.AccountName+"%")
	}

	if err = db.Table(m.Table).
		Preload("MerchantBalances").
		Preload("Users").
		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).Find(&merchants).Error; err != nil {
		return nil, 0, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.Table(m.Table).Count(&count).Error; err != nil {
		return nil, 0, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	return
}

// GetNextMerchantCode 取得最新商戶編號
func (m *Merchant) GetNextMerchantCode() string {
	var code string
	m.MyDB.Table(m.Table).Select("max(code)").Row().Scan(&code)
	reg, _ := regexp.Compile("[^0-9]+")

	if code == "" {
		return "ME00001"
	}
	codeNum, _ := strconv.Atoi(reg.ReplaceAllString(code, ""))
	return "ME" + fmt.Sprintf("%05d", codeNum+1)
}

// GetNextAgentLayerCode 依上級代理 取得最新下級代理層級編號
//func (m *Merchant) GetNextAgentLayerCode(parentMerchant types.Merchant) string {
//	var code string
//
//	m.MyDB.Table(m.Table).Select("max(agent_layer_code)").Where("agent_parent_code = ?", parentMerchant.Code).Row().Scan(&code)
//	reg, _ := regexp.Compile("[^0-9]+")
//
//	if code == "" {
//		return parentMerchant.AgentLayerCode + "001"
//	}
//
//	codeNum, _ := strconv.Atoi(reg.ReplaceAllString(code, ""))
//	return "A" + fmt.Sprintf("%03d", codeNum+1)
//}

func (m *Merchant) GetNextGeneralAgentCode() string {
	var code string
	m.MyDB.Table(m.Table).Select("max(agent_layer_code)").Where("agent_parent_code is null or agent_parent_code = '' ").Row().Scan(&code)
	reg, _ := regexp.Compile("[^0-9]+")
	if code == "" {
		return "A1001"
	}
	codeNum, _ := strconv.Atoi(reg.ReplaceAllString(code, ""))
	return "A" + fmt.Sprintf("%04d", codeNum+1)
}
