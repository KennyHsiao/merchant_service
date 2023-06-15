package model

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"fmt"
	"gorm.io/gorm"
	"regexp"
	"strconv"
)

type AgentRecord struct {
	MyDB  *gorm.DB
	Table string
}

func NewAgentRecord(mydb *gorm.DB, t ...string) *AgentRecord {
	table := "mc_agent_record"
	if len(t) > 0 {
		table = t[0]
	}
	return &AgentRecord{
		MyDB:  mydb,
		Table: table,
	}
}

// GetNextAgentLayerCode 依上級代理 取得最新下級代理層級編號
func (m *AgentRecord) GetNextAgentLayerCode(merchantCode, parentAgentCode, parentAgentLayerCode string) (nextAgentLayerCode string, err error) {
	var code string

	// 若曾經歸於該代理 使用舊層級編號
	m.MyDB.Table(m.Table).
		Select("agent_layer_code").
		Where("merchant_code = ?", merchantCode).
		Where("parent_agent_layer_code = ?", parentAgentLayerCode).
		Row().Scan(&code)

	if len(code) > 0 {
		nextAgentLayerCode = code
		return
	}

	// 若沒有 則依紀錄取得最大代理號
	m.MyDB.Table(m.Table).
		Select("max(agent_layer_code)").
		Where("parent_agent_layer_code = ?", parentAgentLayerCode).
		Row().Scan(&code)

	// 沒有最大代理號則給001
	if code == "" {
		nextAgentLayerCode = parentAgentLayerCode + "001"
	} else {
		// 若有最大代理號 +1
		reg, _ := regexp.Compile("[^0-9]+")
		codeNum, _ := strconv.Atoi(reg.ReplaceAllString(code, ""))
		nextAgentLayerCode = "A" + fmt.Sprintf("%03d", codeNum+1)
	}

	// 新的代理號要記錄
	if err = m.MyDB.Table(m.Table).
		Create(map[string]interface{}{
			"merchant_code":           merchantCode,
			"agent_parent_code":       parentAgentCode,
			"agent_layer_code":        nextAgentLayerCode,
			"parent_agent_layer_code": parentAgentLayerCode,
		}).Error; err != nil {
		return "", errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
