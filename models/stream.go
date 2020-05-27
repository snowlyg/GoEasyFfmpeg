package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/snowlyg/go-rtsp-server/extend/utils"
	"strings"
)

type Stream struct {
	gorm.Model
	URL               string `gorm:"type:varchar(256);primary_key;unique"`
	CustomPath        string `gorm:"type:varchar(256)"`
	OutIp             string `gorm:"type:varchar(256)"`
	HeartbeatInterval int
	Status            bool
}

// GetUrl 获取输出 ip
func (s *Stream) GetUrl() string {
	pathPort := utils.Conf().Section("rtsp").Key("port").MustString("8554")
	ip := s.OutIp

	// 如果自定义 ip 包含协议和端口
	if strings.Contains(s.OutIp, "://") {
		return fmt.Sprintf("%v%v", ip, s.CustomPath)
	}

	// 没有自定义 ip
	if s.OutIp == "" {
		ip = utils.LocalIP()
	}

	return fmt.Sprintf("rtsp://%v:%v%v", ip, pathPort, s.CustomPath)
}
