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

type EncryptionParamLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEncryptionParamLogic(ctx context.Context, svcCtx *svc.ServiceContext) EncryptionParamLogic {
	return EncryptionParamLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EncryptionParamLogic) EncryptionParam(req *types.EncryptionParamRequest) (resp string, err error) {

	var merchant *types.Merchant
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

	// 參數加密
	if respDataBytes, err = utils.AesCBCEncrypt([]byte(req.Data), []byte(merchant.ScrectKey)); err != nil {
		return "", errorz.New(response.API_INVALID_PARAMETER, fmt.Sprintf("CBC加密错误:%s", err.Error()))
	}

	return base64.StdEncoding.EncodeToString(respDataBytes), err
}
