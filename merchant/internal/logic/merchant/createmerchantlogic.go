package merchant

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/random"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"com.copo/bo_service/merchant/internal/model"
	"com.copo/bo_service/merchant/internal/service/merchantsService"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/neccoys/go-zero-extension/redislock"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"time"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) CreateMerchantLogic {
	return CreateMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateMerchantLogic) CreateMerchant(req *types.APIEncryptData) (resp *types.APIEncryptRespData, err error) {
	logx.WithContext(l.ctx).Infof("CreateMerchant enter: %+v", req)

	redisKey := fmt.Sprintf("%s", req.MerchantId)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "create-merchant:")
	redisLock.SetExpire(3)

	if isOK, _ := redisLock.Acquire(); isOK {
		defer redisLock.Release()
		if resp, err = l.DoCreateMerchant(req); err != nil {
			return
		}
	} else {
		return nil, errors.New(response.REDIS_LOCK)
	}

	return
}

func (l *CreateMerchantLogic) DoCreateMerchant(req *types.APIEncryptData) (resp *types.APIEncryptRespData, err error) {

	var merchant *types.Merchant
	var dataBase64 []byte
	var dataBytes []byte
	var respDataBytes []byte
	var autoSignUpMerchantBO types.AutoSignUpMerchantBO

	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", req.MerchantId).
		Where("status = ?", "1").
		Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}
	// 檢查白名單
	if isWhite := merchantsService.IPChecker(req.MyIp, merchant.ApiIP); !isWhite {
		logx.WithContext(l.ctx).Errorf("此IP非法登錄，請設定白名單 来源IP:%s, 白名单:%s", req.MyIp, merchant.ApiIP)
		return nil, errorz.New(response.IP_DENIED, "IP: "+req.MyIp)
	}

	// 檢查商戶狀態
	if merchant.AgentStatus == constants.MerchantAgentStatusDisable ||
		merchant.Status == constants.MerchantStatusDisable ||
		merchant.Status == constants.MerchantStatusClear ||
		len(merchant.AgentLayerCode) == 0 {
		return nil, errorz.New(response.PARENT_MERCHANT_IS_NOT_AGENT, "指定的上層代理，目前不可添加商戶或配置代理 = 代理或該代理已被停權)。")
	}

	// 參數解base64
	if dataBase64, err = base64.StdEncoding.DecodeString(req.Data); err != nil {
		return nil, errorz.New(response.API_INVALID_PARAMETER, fmt.Sprintf("base64解密错误:%s", err.Error()))
	}
	// 參數解密
	if dataBytes, err = utils.AesCBCDecrypt(dataBase64, []byte(merchant.ScrectKey)); err != nil {
		return nil, errorz.New(response.API_INVALID_PARAMETER, fmt.Sprintf("CBC解密错误:%s", err.Error()))
	}
	logx.Infof("request data:%s", string(dataBytes[:]))
	// json to struct
	if err = json.Unmarshal(dataBytes, &autoSignUpMerchantBO); err != nil {
		return nil, errorz.New(response.API_PARAMETER_TYPE_ERROE, fmt.Sprintf("jsonUnmarshal错误:%s", string(dataBytes[:])))
	}
	// 檢查驗簽
	if isSameSign := utils.VerifySign(autoSignUpMerchantBO.Sign, autoSignUpMerchantBO, merchant.ScrectKey, l.ctx); !isSameSign {
		logx.WithContext(l.ctx).Errorf("签名出错")
		return nil, errorz.New(response.INVALID_SIGN, "(sign)签名出错")
	}

	merchantCreateReq := types.MerchantCreateRequest{
		Account: autoSignUpMerchantBO.LoginName,
		Contact: types.MerchantContact{
			Phone:                 autoSignUpMerchantBO.PhoneNumber,
			CommunicationNickname: autoSignUpMerchantBO.CommunicationSoftwareNickname,
			CommunicationSoftware: autoSignUpMerchantBO.CommunicationSoftware,
			Email:                 autoSignUpMerchantBO.Mailbox,
		},
		BizInfo: types.MerchantBizInfo{
			CompanyName: autoSignUpMerchantBO.MerchantCompanyName,
		},
		BoIP:  strings.Join(autoSignUpMerchantBO.LoginIps, ","),
		ApiIP: strings.Join(autoSignUpMerchantBO.ServerIps, ","),
		MerchantCurrencies: []types.MerchantCurrency{
			{
				CurrencyCode: "CNY",
			},
		},
	}
	var newMerchant *types.MerchantCreate

	return resp, l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		if newMerchant, err = l.MerchantCreate(db, merchantCreateReq, merchant.Code, merchant.AgentLayerCode); err != nil {
			return
		}

		if err = l.CreateChangeNotify(db, newMerchant.Code); err != nil {
			return
		}

		respDTO := types.AutoSignUpRespDTO{
			AgentMerchantCoding: autoSignUpMerchantBO.AgentMerchantCoding,
			MappingKey:          autoSignUpMerchantBO.MappingKey,
			MerchantId:          newMerchant.Code,
			MerchantKey:         newMerchant.ScrectKey,
			MerchantApiUrl:      l.svcCtx.Config.Domain,
		}
		respDTO.Sign = utils.SortAndSign2(respDTO, merchant.ScrectKey)

		// struct to json
		respDTOJson, err := json.Marshal(respDTO)
		if err != nil {
			return errorz.New(response.API_PARAMETER_TYPE_ERROE, fmt.Sprintf("jsonMarshal错误:%+v", respDTOJson))
		}
		// 參數加密
		if respDataBytes, err = utils.AesCBCEncrypt(respDTOJson, []byte(merchant.ScrectKey)); err != nil {
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

func (l *CreateMerchantLogic) MerchantCreate(db *gorm.DB, req types.MerchantCreateRequest, agentParentCode, parentAgentLayerCode string) (merchant *types.MerchantCreate, err error) {
	merchant = &types.MerchantCreate{
		MerchantCreateRequest: req,
	}
	l.SetInitialValue(merchant)

	if err = l.Verify(db, merchant); err != nil {
		return
	}

	mux := model.NewMerchant(db)
	merchant.Code = mux.GetNextMerchantCode()

	password := random.GetRandomString(10, random.ALL, random.MIX)
	_, err1 := l.userCreate(db, merchant, password)
	if err1 != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err1.Error())
	}

	mcux := model.NewMerchantCurrency(db)
	mbux := model.NewMerchantBalance(db)

	for _, currency := range req.MerchantCurrencies {
		if err = mcux.CreateMerchantCurrency(merchant.Code, currency.CurrencyCode, "1"); err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		if err = mbux.CreateMerchantBalances(merchant.Code, currency.CurrencyCode); err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	var agentLayerCode string

	if agentLayerCode, err = model.NewAgentRecord(db).GetNextAgentLayerCode(merchant.Code, agentParentCode, parentAgentLayerCode); err != nil {
		return nil, errorz.New(response.CHANGE_AGENT_ERROR, err.Error())
	}

	merchant.AgentLayerCode = agentLayerCode
	merchant.AgentParentCode = agentParentCode

	if err = db.Table("mc_merchants").
		Omit("Users.*").Omit("MerchantCurrencies").
		Create(merchant).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func (l *CreateMerchantLogic) SetInitialValue(merchant *types.MerchantCreate) {

	merchant.ScrectKey = random.GetRandomString(32, random.ALL, random.MIX)
	merchant.AccountName = merchant.Account
	merchant.Status = constants.MerchantStatusEnable
	merchant.AgentStatus = constants.MerchantAgentStatusDisable
	merchant.LoginValidatedType = "1"
	merchant.PayingValidatedType = "1"
	merchant.ApiCodeType = "1"
	merchant.BillLadingType = "0"
	merchant.Lang = "CN"
	merchant.RegisteredAt = time.Now().Unix()
	merchant.Users = append(merchant.Users, types.User{Account: merchant.Account})
}

func (l *CreateMerchantLogic) Verify(db *gorm.DB, merchant *types.MerchantCreate) (err error) {
	var isExist bool

	if isExist, err = model.NewUser(db).IsExistByAccount(merchant.Account); err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	} else if isExist {
		return errorz.New(response.USER_HAS_REGISTERED, "该用户名已存在")
	}

	if isExist, err = model.NewUser(db).IsExistByEmail(merchant.Contact.Email); err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	} else if isExist {
		return errorz.New(response.MAILBOX_HAS_REGISTERED, "该邮箱已存在")
	}

	var ips []string

	if len(merchant.BoIP) > 0 {
		ips = append(ips, strings.Split(merchant.BoIP, ",")...)
	}
	if len(merchant.ApiIP) > 0 {
		ips = append(ips, strings.Split(merchant.ApiIP, ",")...)
	}

	for _, ip := range ips {
		if isMatch, _ := regexp.MatchString(constants.RegexpIpaddressPattern, ip); !isMatch {
			return errorz.New(response.ILLEGAL_IP, "IP格式错误")
		}
	}

	return
}

func (l *CreateMerchantLogic) userCreate(db *gorm.DB, merchant *types.MerchantCreate, password string) (*types.UserCreate, error) {

	var currencies []types.Currency

	for _, currency := range merchant.MerchantCurrencies {
		currencies = append(currencies, types.Currency{
			Code: currency.CurrencyCode,
		})
	}
	user := &types.UserCreate{
		UserCreateRequest: types.UserCreateRequest{
			Account:      merchant.Account,
			Name:         merchant.Account,
			Email:        merchant.Contact.Email,
			RegisteredAt: time.Now().Unix(),
			Roles: []types.Role{{
				ID: 2,
			}},
			Password:      utils.PasswordHash2(password),
			Currencies:    currencies,
			DisableDelete: "1",
			Status:        "1",
			IsLogin:       "0",
			IsAdmin:       "0",
		},
	}

	return user, db.Table("au_users").
		Omit("Merchants.*").
		Omit("Roles.*").Create(user).Error
}

func (l *CreateMerchantLogic) CreateChangeNotify(db *gorm.DB, merchantCode string) (err error) {
	var systemParam types.SystemParams

	if err = db.Table("bs_system_params").Where("name = 'IPH-paymentlist-notify'").Take(&systemParam).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, fmt.Sprintf("取得通知URL失败:%s", err.Error()))
	}

	if err = db.Table("mc_channel_change_notify").Create(&types.ChannelChangeNotify{
		MerchantCode:          merchantCode,
		IsChannelChangeNotify: "1",
		NotifyUrl:             systemParam.Value,
		Status:                "1",
		CreatedBy:             "admin",
		UpdatedBy:             "admin",
	}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, fmt.Sprintf("设定通知URL失败:%s", err.Error()))
	}

	return
}
