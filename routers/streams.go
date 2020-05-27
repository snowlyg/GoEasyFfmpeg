package routers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/snowlyg/go-rtsp-server/extend/EasyGoLib/db"
	"github.com/snowlyg/go-rtsp-server/models"
	"github.com/snowlyg/go-rtsp-server/rtsp"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/**
 * @apiDefine stream 流管理
 */

/**
 * @api {get} /api/v1/stream/add 启动拉转推
 * @apiGroup stream
 * @apiName StreamAdd
 * @apiParam {String} url RTSP源地址
 * @apiParam {String} [customPath] 转推时的推送PATH
 * @apiParam {String=TCP,UDP} [transType=TCP] 拉流传输模式
 * @apiParam {Number} [idleTimeout] 拉流时的超时时间
 * @apiParam {Number} [heartbeatInterval] 拉流时的心跳间隔，毫秒为单位。如果心跳间隔不为0，那拉流时会向源地址以该间隔发送OPTION请求用来心跳保活
 * @apiSuccess (200) {String} ID	拉流的ID。后续可以通过该ID来停止拉流
 */
func (h *APIHandler) StreamAdd(c *gin.Context) {
	type Form struct {
		Id                uint   `form:"id" `
		URL               string `form:"source" binding:"required"`
		CustomPath        string `form:"customPath"`
		TransType         string `form:"transType"`
		TransRtpType      string `form:"transRtpType"`
		IdleTimeout       int    `form:"idleTimeout"`
		HeartbeatInterval int    `form:"heartbeatInterval"`
	}
	var form Form
	err := c.Bind(&form)
	if err != nil {
		log.Printf("Pull to push err:%v", err)
		return
	}

	transType := 0
	if form.TransType == "TCP" {
		transType = 0
	} else if form.TransType == "UDP" {
		transType = 1
	}

	// save to db.
	oldStream := models.Stream{}
	if db.SQLite.Where("id = ? ", form.Id).First(&oldStream).RecordNotFound() {

		stream := models.Stream{
			URL:               form.URL,
			CustomPath:        form.CustomPath,
			IdleTimeout:       form.IdleTimeout,
			TransType:         transType,
			TransRtpType:      form.TransRtpType,
			HeartbeatInterval: form.HeartbeatInterval,
			Status:            false,
		}
		db.SQLite.Create(&stream)
		c.IndentedJSON(200, stream)
	} else {
		oldStream.URL = form.URL
		oldStream.CustomPath = form.CustomPath
		oldStream.TransType = transType
		oldStream.TransRtpType = form.TransRtpType
		oldStream.IdleTimeout = form.IdleTimeout
		oldStream.HeartbeatInterval = form.HeartbeatInterval
		oldStream.Status = false
		db.SQLite.Save(oldStream)
		c.IndentedJSON(200, oldStream)
	}

}

/**
 * @api {get} /api/v1/stream/stop 停止推流
 * @apiGroup stream
 * @apiName StreamStop
 * @apiParam {String} id 拉流的ID
 * @apiUse simpleSuccess
 */
func (h *APIHandler) StreamStop(c *gin.Context) {

	type Form struct {
		ID string `form:"id" binding:"required"`
	}

	var form Form
	err := c.Bind(&form)
	if err != nil {
		log.Printf("stop pull to push err:%v", err)
		return
	}

	stream := getStream(form.ID)
	pushers := rtsp.GetServer().GetPushers()
	for _, v := range pushers {
		//if v.URL() == stream.URL {
		v.Stop()
		rtsp.GetServer().RemovePusher(v)
		c.IndentedJSON(200, "OK")
		log.Printf("Stop %v success ", v)
		//if v.RTSPClient != nil {
		stream.Status = false
		stream.StreamId = ""
		db.SQLite.Save(stream)
		//}
		return
		//}
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Pusher[%s] not found", stream.StreamId))
}

func getStream(formId string) models.Stream {
	id, _ := strconv.ParseUint(formId, 10, 64)
	stream := models.Stream{}
	db.SQLite.Where("id = ?", id).First(&stream)
	return stream
}

/**
 * @api {get} /api/v1/stream/start 启动推流
 * @apiGroup stream
 * @apiName StreamStart
 * @apiParam {String} id 拉流的ID
 * @apiUse simpleSuccess
 */
func (h *APIHandler) StreamStart(c *gin.Context) {

	type Form struct {
		ID string `form:"id" binding:"required"`
	}

	var form Form
	err := c.Bind(&form)
	if err != nil {
		log.Printf("stop pull to push err:%v", err)
		return
	}

	stream := getStream(form.ID)
	agent := fmt.Sprintf("EasyDarwinGo/%s", BuildVersion)
	if BuildDateTime != "" {
		agent = fmt.Sprintf("%s(%s)", agent, BuildDateTime)
	}

	p := rtsp.GetServer().GetPusher(stream.URL)
	if p != nil {
		rtsp.GetServer().RemovePusher(p)
	}

	pusher := rtsp.NewClientPusher()

	log.Printf("Pull to push %v success ", stream.StreamId)
	rtsp.GetServer().AddPusher(pusher)

	//if pusher.RTSPClient != nil && !pusher.Stoped() {
	stream.StreamId = pusher.ID()
	stream.Status = true
	db.SQLite.Save(stream)
	c.IndentedJSON(200, "OK")
	log.Printf("Start %v success ", pusher)
	return
	//}

	c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Pusher[%s] not found or not start", form.ID))
}

/**
 * @api {get} /api/v1/stream/del 删除推流
 * @apiGroup stream
 * @apiName StreamDel
 * @apiParam {String} id 拉流的ID
 * @apiUse simpleSuccess
 */
func (h *APIHandler) StreamDel(c *gin.Context) {

	type Form struct {
		ID string `form:"id" binding:"required"`
	}

	var form Form
	err := c.Bind(&form)
	if err != nil {
		log.Printf("stop pull to push err:%v", err)
		return
	}

	stream := getStream(form.ID)

	db.SQLite.Unscoped().Delete(stream)

}
