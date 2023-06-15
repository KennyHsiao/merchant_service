package test

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"gorm.io/gorm"

	"com.copo/bo_service/merchant/internal/svc"
	"com.copo/bo_service/merchant/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DecryptParamLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDecryptParamLogic(ctx context.Context, svcCtx *svc.ServiceContext) DecryptParamLogic {
	return DecryptParamLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DecryptParamLogic) DecryptParam(req *types.EncryptionParamRequest) (resp string, err error) {

	var merchant *types.Merchant
	var respDataBase64 []byte
	var respDataBytes []byte
	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", req.MerchantId).
		Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			return "", errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}
	// 參數解base64
	if respDataBase64, err = base64.StdEncoding.DecodeString(req.Data); err != nil {
		return "", errorz.New(response.API_INVALID_PARAMETER, fmt.Sprintf("base64解密错误:%s", err.Error()))
	}
	// 參數解密
	if respDataBytes, err = utils.AesCBCDecrypt(respDataBase64, []byte(merchant.ScrectKey)); err != nil {
		return "", errorz.New(response.API_INVALID_PARAMETER, fmt.Sprintf("CBC解密错误:%s", err.Error()))
	}

	return string(respDataBytes[:]), err
}
