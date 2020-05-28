package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/snowlyg/go-rtsp-server/extend/EasyGoLib/db"

	figure "github.com/common-nighthawk/go-figure"
	"github.com/kardianos/service"
	"github.com/snowlyg/go-rtsp-server/extend/EasyGoLib/utils"
	"github.com/snowlyg/go-rtsp-server/models"
	"github.com/snowlyg/go-rtsp-server/routers"
	"github.com/snowlyg/go-rtsp-server/rtsp"
)

var (
	gitCommitCode string
	buildDateTime string
)

type program struct {
	httpPort   int
	httpServer *http.Server
	//rtspPort   int
	rtspServer *rtsp.Server
}

// StopHTTP 停止 http
func (p *program) StopHTTP() (err error) {
	if p.httpServer == nil {
		err = fmt.Errorf("HTTP Server Not Found")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = p.httpServer.Shutdown(ctx); err != nil {
		return
	}
	return
}

// StartHTTP 启动 http
func (p *program) StartHTTP() (err error) {
	p.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", p.httpPort),
		Handler:           routers.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	link := fmt.Sprintf("http://%s:%d", utils.LocalIP(), p.httpPort)
	log.Println("http server start -->", link)
	go func() {
		if err := p.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("start http server error", err)
		}
		log.Println("http server start")
	}()
	return
}

// StartRTSP 启动 rtsp
func (p *program) StartRTSP() (err error) {
	if p.rtspServer == nil {
		err = fmt.Errorf("RTSP Server Not Found")
		return
	}
	go func() {
		if err := p.rtspServer.Start(); err != nil {
			log.Println("start rtsp server error", err)
		}
		log.Println("rtsp server start")
	}()
	return
}

// StopRTSP 停止 rtsp
func (p *program) StopRTSP() (err error) {
	if p.rtspServer == nil {
		err = fmt.Errorf("RTSP Server Not Found")
		return
	}
	p.rtspServer.Stop()
	return
}

// Start 启动服务
func (p *program) Start(s service.Service) (err error) {

	log.Println("********** START **********")
	if utils.IsPortInUse(p.httpPort) {
		err = fmt.Errorf("HTTP port[%d] In Use", p.httpPort)
		return
	}

	// 初始化数据库和模型
	err = models.Init()
	if err != nil {
		return
	}
	// 初始化路由
	err = routers.Init()
	if err != nil {
		return
	}

	_ = p.StartHTTP()
	_ = p.StartRTSP()

	if !utils.Debug {
		log.Println("log files -->", utils.LogDir())
		log.SetOutput(utils.GetLogWriter())
	}

	go func() {
		for range routers.API.RestartChan {
			_ = p.StopHTTP()
			_ = p.StopRTSP()
			// 重载配置
			utils.ReloadConf()
			_ = p.StartHTTP()
			_ = p.StartRTSP()
		}
	}()

	go func() {
		log.Printf("demon pull streams")
		for {
			var streams []models.Stream
			if err := db.SQLite.Find(&streams).Error; err != nil {
				log.Printf("find stream err:%v", err)
				return
			}

			for i := len(streams) - 1; i > -1; i-- {
				v := streams[i]
				pusher := rtsp.NewClientPusher(v.ID, v.URL, v.CustomPath)
				if rtsp.GetServer().GetPusher(v.CustomPath) != nil {
					continue
				}
				if v.Status {
					pusher.Stoped = false
					rtsp.GetServer().AddPusher(pusher)
				}
				//streams = streams[0:i]
				//streams = append(streams[:i], streams[i+1:]...)
			}
			time.Sleep(1 * time.Second)
		}
	}()
	log.Printf("server start ")
	return
}

// Stop 停止服务
func (p *program) Stop(s service.Service) (err error) {
	defer log.Println("********** STOP **********")
	defer utils.CloseLogWriter()
	_ = p.StopHTTP()
	_ = p.StopRTSP()
	models.Close()
	return
}

func main() {

	flag.StringVar(&utils.FlagVarConfFile, "config", "", "configure file path")
	flag.Parse()
	tail := flag.Args()

	// log
	log.SetPrefix("[EasyDarwin] ")
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	log.Printf("git commit code:%s", gitCommitCode)
	log.Printf("build date:%s", buildDateTime)

	routers.BuildVersion = fmt.Sprintf("%s.%s", routers.BuildVersion, gitCommitCode)
	routers.BuildDateTime = buildDateTime

	sec := utils.Conf().Section("service")
	svcConfig := &service.Config{
		Name:        sec.Key("name").MustString("EasyDarwin_Service"),
		DisplayName: sec.Key("display_name").MustString("EasyDarwin_Service"),
		Description: sec.Key("description").MustString("EasyDarwin_Service"),
	}

	httpPort := utils.Conf().Section("http").Key("port").MustInt(10008)
	rtspServer := rtsp.GetServer()
	p := &program{
		httpPort:   httpPort,
		rtspServer: rtspServer,
	}
	s, err := service.New(p, svcConfig)
	if err != nil {
		log.Println(err)
		utils.PauseExit()
	}
	if len(tail) > 0 {
		cmd := strings.ToLower(tail[0])
		if cmd == "install" || cmd == "stop" || cmd == "start" || cmd == "uninstall" {
			figure.NewFigure("EasyDarwin", "", false).Print()
			log.Println(svcConfig.Name, cmd, "...")
			if err = service.Control(s, cmd); err != nil {
				log.Println(err)
				utils.PauseExit()
			}
			log.Println(svcConfig.Name, cmd, "ok")
			return
		}
	}
	figure.NewFigure("EasyDarwin", "", false).Print()
	if err = s.Run(); err != nil {
		log.Println(err)
		utils.PauseExit()
	}
}
