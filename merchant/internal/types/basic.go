package types

import (
	"com.copo/bo_service/common/gormx"
	"time"
)

func (Currency) TableName() string {
	return "bs_currencies"
}

type TimezonexCreate struct {
	TimezonexCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TimezonexUpdate struct {
	TimezonexUpdateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CurrencyQueryAllRequestX struct {
	CurrencyQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type CurrencyCreate struct {
	CurrencyCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CurrencyUpdate struct {
	CurrencyUpdateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LanguageCreate struct {
	LanguageCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LanguageUpdate struct {
	LanguageUpdateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SystemRateCeate struct {
	SystemRate
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SystemParams struct {
	Name  string
	Type  string
	Value string
	Title string
}
