package autoFillBankIdService

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/merchant/internal/types"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strings"
)

// AutoFillBankId 22/08/31  代付提單沒給bankId, 用銀行名稱判斷補給 (目前僅支援MLB代付)
func AutoFillBankId(ctx context.Context, db *gorm.DB, req *types.ProxyPayRequestX) (err error) {

	bankName := req.BankName
	bankNameChars := []rune(bankName)
	var likeNames []string

	if len(bankName) == 0 {
		logx.WithContext(ctx).Errorf("AutoFillBankId error: bankName is empty")
		return errorz.New(response.INVALID_BANK_ID, "BankName: "+req.BankName)
	}

	// 先判斷是否為 农村信用社 或 农村商业银行
	if strings.Index(bankName, "信用社") >= 0 {
		req.BankId = "402" // 农村信用社
	} else if strings.Index(bankName, "农信社") >= 0 {
		req.BankId = "402" // 农村信用社
	} else if strings.Index(bankName, "信用合作社") >= 0 {
		req.BankId = "402" // 农村信用社
	} else if strings.Index(bankName, "农村商业") >= 0 {
		req.BankId = "314" // 农村商业银行
	} else if strings.Index(bankName, "农商") >= 0 {
		req.BankId = "314" // 农村商业银行
	}

	if req.BankId != "" {
		logx.WithContext(ctx).Infof("AutoFillBankId ok: bankName:%s,bankId:%s", bankName, req.BankId)
		return nil
	}

	//1. %中国工银行%      = 前後+%
	likeNames = append(likeNames, "%"+bankName+"%")

	//2. %中%国%工%银%行%  = 字間頭尾+%
	likeNameA := ""
	for _, char := range bankNameChars {
		likeNameA += "%" + string(char) + "%"
	}
	likeNames = append(likeNames, likeNameA)

	if strings.Index(bankName, "中国") >= 0 {
		newBankName := strings.Replace(bankName, "中国", "", 1)
		newBankNameChars := []rune(newBankName)

		//3. %工银行%         = 去掉"中国"後前後+%
		likeNames = append(likeNames, "%"+newBankName+"%")

		//4. %工%银%行%       = 去掉"中国"後字間頭尾+%
		likeNameB := ""
		for _, char := range newBankNameChars {
			likeNameB += "%" + string(char) + "%"
		}
		likeNames = append(likeNames, likeNameB)
	}

	//找出多組比對字數 多組時在搜尋字數少  字數最少還是有多組就報錯
	for _, likeName := range likeNames {
		isOk, bankId, errX := findByLikeBankName(ctx, db, req.Currency, likeName)
		if errX != nil {
			return errX
		} else if isOk {
			req.BankId = bankId
			logx.WithContext(ctx).Infof("AutoFillBankId ok: bankName:%s,bankId:%s", bankName, req.BankId)
			return nil
		}
	}
	logx.WithContext(ctx).Errorf("AutoFillBankId error: 找不到符合銀行 BankName: %s", bankName)
	return errorz.New(response.INVALID_BANK_ID, "AutoFillBankId error: 找不到符合銀行 BankName: "+bankName)
}

func findByLikeBankName(ctx context.Context, db *gorm.DB, currencyCode, bankName string) (isOk bool, bankId string, err error) {
	var banks []types.Bank
	if err = db.Table("bk_banks").
		Where("bank_name like ? ", bankName).
		Where("currency_code = ? ", currencyCode).Find(&banks).Error; err != nil {
		return
	}

	if len(banks) == 0 {
		// 找不到對應表示 [失敗不抱錯]
		return
	} else if len(banks) == 1 {
		// 找到一個對應表示 [完成]
		bankId = banks[0].BankNo
		isOk = true
		return
	} else {
		// 找到多個對應 要比對字數 若沒有唯一字數最少的 [報錯]
		var selectedBank types.Bank
		wordCount := 99       // 字數
		currentRepeatNum := 0 // 該字數重複次數

		for _, bank := range banks {
			// 當找到字數更少時
			if wordCount > len(bank.BankName) {
				selectedBank = bank
				wordCount = len(bank.BankName)
				currentRepeatNum = 0
			} else if wordCount == len(bank.BankName) { // 當找到字數相同時
				currentRepeatNum += 1
			}

		}

		if currentRepeatNum == 0 {
			isOk = true
			bankId = selectedBank.BankNo
			return
		} else {
			logx.WithContext(ctx).Errorf("找到多種符合銀行 BankName: %s", bankName)
			return false, "", errorz.New(response.INVALID_BANK_ID, "找到多種對應 BankName: "+bankName)
		}

	}

}
