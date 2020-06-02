package models

import (
	"github.com/jinzhu/gorm"
)

type Stream struct {
	gorm.Model
	Status     bool
	URL        string `gorm:"type:varchar(256);unique"`
	CustomPath string `gorm:"type:varchar(256)"`
	TransType  string `gorm:"type:varchar(256)"`
	RoomName   string `gorm:"type:varchar(256)"`
}
