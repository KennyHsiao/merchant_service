package merchant

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/random"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"gorm.io/gorm"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) QueryMerchantLogic {
	return QueryMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMerchantLogic) QueryMerchant(req *types.APIEncryptData) (resp *types.APIEncryptRespData, err error) {

	var parentMerchant *types.Merchant
	var merchant *types.Merchant
	var user *types.User
	var dataBase64 []byte
	var dataBytes []byte
	var respDataBytes []byte
	var queryMerchantBO types.QueryMerchantBO

	// 取得父商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", req.MerchantId).
		Where("status = ?", "1").
		Take(&parentMerchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	// 檢查白名單
	if isWhite := merchantsService.IPChecker(req.MyIp, parentMerchant.ApiIP); !isWhite {
		logx.WithContext(l.ctx).Errorf("此IP非法登錄，請設定白名單 来源IP:%s, 白名单:%s", req.MyIp, parentMerchant.ApiIP)
		return nil, errorz.New(response.IP_DENIED, "IP: "+req.MyIp)
	}

	// 參數解base64
	if dataBase64, err = base64.StdEncoding.DecodeString(req.Data); err != nil {
		return nil, errorz.New(response.API_INVALID_PARAMETER, fmt.Sprintf("base64解密错误:%s", err.Error()))
	}
	// 參數解密
	if dataBytes, err = utils.AesCBCDecrypt(dataBase64, []byte(parentMerchant.ScrectKey)); err != nil {
		return nil, errorz.New(response.API_INVALID_PARAMETER, fmt.Sprintf("CBC解密错误:%s", err.Error()))
	}
	logx.Infof("request data:%s", string(dataBytes[:]))
	// json to struct
	if err = json.Unmarshal(dataBytes, &queryMerchantBO); err != nil {
		return nil, errorz.New(response.API_PARAMETER_TYPE_ERROE, fmt.Sprintf("jsonUnmarshal错误:%s", string(dataBytes[:])))
	}
	// 檢查驗簽
	if isSameSign := utils.VerifySign(queryMerchantBO.Sign, queryMerchantBO, parentMerchant.ScrectKey, l.ctx); !isSameSign {
		logx.WithContext(l.ctx).Errorf("签名出错")
		return nil, errorz.New(response.INVALID_SIGN, "(sign)签名出错")
	}

	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", queryMerchantBO.MerchantId).
		Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	if merchant.AgentParentCode != parentMerchant.Code {
		logx.WithContext(l.ctx).Errorf("代理%s,商户%s不是代理关系:%s,%s", parentMerchant.Code, merchant.Code, parentMerchant.AgentLayerCode, merchant.AgentLayerCode)
		return nil, errorz.New(response.INVALID_MERCHANT_AGENT, fmt.Sprintf("代理%s,商户%s不是代理关系:%s,%s", parentMerchant.Code, merchant.Code, parentMerchant.AgentLayerCode, merchant.AgentLayerCode))
	}

	// 取得帳號
	if err = l.svcCtx.MyDB.Table("au_users").
		Where("account = ?", merchant.AccountName).
		Take(&user).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, fmt.Sprintf("取得帐号失败:%s", err.Error()))
	}

	return resp, l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {

		screctKey := random.GetRandomString(32, random.ALL, random.MIX)
		password := random.GetRandomString(10, random.ALL, random.MIX)
		shaPassword := utils.PasswordHash2(password)

		merchant.ScrectKey = screctKey
		user.Password = shaPassword

		// 变更密码
		if err = db.Table("au_users").Updates(user).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, fmt.Sprintf("变更密码失败:%s", err.Error()))
		}
		// 变更密钥
		if err = db.Table("mc_merchants").Updates(merchant).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, fmt.Sprintf("变更密钥失败:%s", err.Error()))
		}

		respDTO := types.QueryMerchantDTO{
			MerchantId:                    merchant.Code,
			LoginName:                     merchant.AccountName,
			Mailbox:                       merchant.Contact.Email,
			PhoneNumber:                   merchant.Contact.Phone,
			CommunicationSoftware:         merchant.Contact.CommunicationSoftware,
			CommunicationSoftwareNickname: merchant.Contact.CommunicationNickname,
			LoginIps:                      merchant.BoIP,
			MerchantCompanyName:           merchant.BizInfo.CompanyName,
			Password:                      password,
			MerchantOperatingWebsite:      merchant.BizInfo.OperatingWebsite,
			MerchantTestAccount:           "",
			MerchantTestPassword:          "",
			MerchantKey:                   screctKey,
		}
		respDTO.Sign = utils.SortAndSign2(respDTO, parentMerchant.ScrectKey)

		// struct to json
		respDTOJson, err := json.Marshal(respDTO)
		if err != nil {
			return errorz.New(response.API_PARAMETER_TYPE_ERROE, fmt.Sprintf("jsonMarshal错误:%+v", respDTOJson))
		}
		// 參數加密
		if respDataBytes, err = utils.AesCBCEncrypt(respDTOJson, []byte(parentMerchant.ScrectKey)); err != nil {
			return errorz.New(response.API_INVALID_PARAMETER, fmt.Sprintf("CBC加密错误:%s", err.Error()))
		}

		// 參數base64加密
		resp = &types.APIEncryptRespData{
			Data:       base64.StdEncoding.EncodeToString(respDataBytes),
			MerchantId: req.MerchantId,
			Version:    "1.0",
			RespCode:   response.API_SUCCESS,
			RespMsg:    i18n.Sprintf(response.API_SUCCESS),
		}
		return
	})
}
