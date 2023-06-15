package model

import (
	"gorm.io/gorm"
)

type user struct {
	MyDB  *gorm.DB
	Table string
}

func NewUser(mydb *gorm.DB, t ...string) *user {
	table := "au_users"
	if len(t) > 0 {
		table = t[0]
	}
	return &user{
		MyDB:  mydb,
		Table: table,
	}
}

func (u *user) IsExistByAccount(account string) (isExist bool, err error) {
	err = u.MyDB.Table(u.Table).
		Select("count(*) > 0").
		Where("account = ?", account).
		Find(&isExist).Error
	return
}

func (u *user) IsExistByEmail(email string) (isExist bool, err error) {
	err = u.MyDB.Table(u.Table).
		Select("count(*) > 0").
		Where("email = ?", email).
		Find(&isExist).Error
	return
}
