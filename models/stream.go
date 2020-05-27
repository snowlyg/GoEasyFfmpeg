package models

import (
	"github.com/jinzhu/gorm"
)

type Stream struct {
	gorm.Model
	URL               string `gorm:"type:varchar(256);primary_key;unique"`
	CustomPath        string `gorm:"type:varchar(256)"`
	OutIp             string `gorm:"type:varchar(256)"`
	HeartbeatInterval int
	Status            bool
}
