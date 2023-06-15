package types

import (
	"com.copo/bo_service/common/gormx"
	"time"
)

func (User) TableName() string {
	return "au_users"
}

func (Role) TableName() string {
	return "au_roles"
}

func (Permit) TableName() string {
	return "au_permits"
}

func (Menu) TableName() string {
	return "au_menus"
}

type MenuCreate struct {
	MenuCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MenuUpdate struct {
	MenuUpdateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserX struct {
	User
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserCreate struct {
	UserCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserUpdate struct {
	UserUpdateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PermitCreate struct {
	PermitCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PermitUpdate struct {
	PermitUpdateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RoleCreate struct {
	RoleCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RoleUpdate struct {
	RoleUpdateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LoginResponseX struct {
	ID           int64       `json:"id"`
	Account      string      `json:"account"`
	IsAdmin      bool        `json:"isAdmin"`
	Identity     string      `json:"identity"`
	MerchantCode string      `json:"merchantCode"`
	Jwt          JwtToken    `json:"jwt"`
	MenuTree     []*MenuTree `json:"menuTree"`
}

type UserMenuResponseX struct {
	MenuTree []*MenuTree `json:"menuTree"`
}

type MerchantUserQueryAllRequestX struct {
	MerchantUserQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type MenuQueryAllRequestX struct {
	MenuQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}
