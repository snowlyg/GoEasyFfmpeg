package models

import (
	"github.com/jinzhu/gorm"
)

type Stream struct {
	gorm.Model
	URL               string `gorm:"type:varchar(256);primary_key;unique"`
	StreamId          string `gorm:"type:varchar(256)"`
	CustomPath        string `gorm:"type:varchar(256)"`
	TransType         int
	TransRtpType      string `gorm:"type:varchar(256)"`
	IdleTimeout       int
	HeartbeatInterval int
	Status            bool
}
