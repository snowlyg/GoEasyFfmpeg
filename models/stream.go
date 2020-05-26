package models

import (
	"github.com/jinzhu/gorm"
)

type Stream struct {
	gorm.Model
	URL               string `gorm:"type:varchar(256);primary_key;unique"`
	CustomPath        string `gorm:"type:varchar(256)"`
	IdleTimeout       int
	HeartbeatInterval int
	Status            bool
}
