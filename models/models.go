package models

import (
	"github.com/snowlyg/GoEasyFfmpeg/extend/EasyGoLib/db"
	"github.com/snowlyg/GoEasyFfmpeg/extend/EasyGoLib/utils"
)

func Init() (err error) {
	err = db.Init()
	if err != nil {
		return
	}
	db.SQLite.AutoMigrate(&User{}, &Stream{})
	count := 0
	sec := utils.Conf().Section("http")
	defUser := sec.Key("default_username").MustString("admin")
	defPass := sec.Key("default_password").MustString("admin")
	db.SQLite.Model(User{}).Where("username = ?", defUser).Count(&count)
	if count == 0 {
		db.SQLite.Create(&User{
			Username: defUser,
			Password: utils.MD5(defPass),
		})
	}
	return
}

func Close() {
	db.Close()
}
