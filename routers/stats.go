package routers

import (
	"fmt"
	"github.com/snowlyg/go-rtsp-server/extend/db"
	"github.com/snowlyg/go-rtsp-server/models"
	"log"
	//"strings"

	"github.com/gin-gonic/gin"
	"github.com/snowlyg/go-rtsp-server/extend/utils"
)

/**
 * @apiDefine stats 统计
 */

/**
 * @apiDefine playerInfo
 * @apiSuccess (200) {String} rows.id
 * @apiSuccess (200) {String} rows.path
 * @apiSuccess (200) {String} rows.transType 传输模式
 * @apiSuccess (200) {Number} rows.inBytes 入口流量
 * @apiSuccess (200) {Number} rows.outBytes 出口流量
 * @apiSuccess (200) {String} rows.startAt 开始时间
 */

/**
 * @api {get} /api/v1/pushers 获取推流列表
 * @apiGroup stats
 * @apiName Pushers
 * @apiParam {Number} [start] 分页开始,从零开始
 * @apiParam {Number} [limit] 分页大小
 * @apiParam {String} [sort] 排序字段
 * @apiParam {String=ascending,descending} [order] 排序顺序
 * @apiParam {String} [q] 查询参数
 * @apiSuccess (200) {Number} total 总数
 * @apiSuccess (200) {Array} rows 推流列表
 * @apiSuccess (200) {String} rows.id
 * @apiSuccess (200) {String} rows.streamId
 * @apiSuccess (200) {String} rows.path
 * @apiSuccess (200) {String} rows.transType 传输模式
 * @apiSuccess (200) {Number} rows.inBytes 入口流量
 * @apiSuccess (200) {Number} rows.outBytes 出口流量
 * @apiSuccess (200) {String} rows.startAt 开始时间
 * @apiSuccess (200) {Number} rows.onlines 在线人数
 */
func (h *APIHandler) Pushers(c *gin.Context) {

	form := utils.NewPageForm()
	if err := c.Bind(form); err != nil {
		return
	}

	//hostname := utils.GetRequestHostname(c.Request)
	var streams []models.Stream
	if err := db.SQLite.Find(&streams).Error; err != nil {
		log.Printf("find stream err:%v", err)
		return
	}
	pathIp := utils.Conf().Section("rtsp").Key("port").MustString("8554")
	pushers := make([]interface{}, 0)
	for _, stream := range streams {
		statusText := "已停止"
		if stream.Status {
			statusText = "已启动"
		}
		url := fmt.Sprintf("rtsp://%v:%v%v", utils.LocalIP(), pathIp, stream.CustomPath)
		pushers = append(pushers, map[string]interface{}{
			"id":         stream.ID,
			"source":     stream.URL,
			"customPath": stream.CustomPath,
			"outIp":      stream.OutIp,
			"url":        url,
			"status":     statusText,
		})
	}

	pr := utils.NewPageResult(pushers)
	if form.Sort != "" {
		pr.Sort(form.Sort, form.Order)
	}

	pr.Slice(form.Start, form.Limit)
	c.IndentedJSON(200, pr)
}

/**
 * @api {get} /api/v1/players 获取拉流列表
 * @apiGroup stats
 * @apiName Players
 * @apiParam {Number} [start] 分页开始,从零开始
 * @apiParam {Number} [limit] 分页大小
 * @apiParam {String} [sort] 排序字段
 * @apiParam {String=ascending,descending} [order] 排序顺序
 * @apiParam {String} [q] 查询参数
 * @apiSuccess (200) {Number} total 总数
 * @apiSuccess (200) {Array} rows 推流列表
 * @apiSuccess (200) {String} rows.id
 * @apiSuccess (200) {String} rows.path
 * @apiSuccess (200) {String} rows.transType 传输模式
 * @apiSuccess (200) {Number} rows.inBytes 入口流量
 * @apiSuccess (200) {Number} rows.outBytes 出口流量
 * @apiSuccess (200) {String} rows.startAt 开始时间
 */
func (h *APIHandler) Players(c *gin.Context) {
	form := utils.NewPageForm()
	if err := c.Bind(form); err != nil {
		return
	}

	//hostname := utils.GetRequestHostname(c.Request)
	_players := make([]interface{}, 0)

	_players = append(_players, map[string]interface{}{})

	pr := utils.NewPageResult(_players)
	if form.Sort != "" {
		pr.Sort(form.Sort, form.Order)
	}
	pr.Slice(form.Start, form.Limit)
	c.IndentedJSON(200, pr)
}
