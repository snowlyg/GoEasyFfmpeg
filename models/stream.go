package models

import (
	"github.com/jinzhu/gorm"
)

type Stream struct {
	gorm.Model
	URL        string `gorm:"type:varchar(256);primary_key;unique"`
	CustomPath string `gorm:"type:varchar(256)"`
	TransType  string `gorm:"type:varchar(256)"`
	RoomName   string `gorm:"type:varchar(256)"`
	RoomKey    string `gorm:"type:varchar(256)"`
	Status     bool
}
