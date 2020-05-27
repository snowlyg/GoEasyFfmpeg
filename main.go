package main

import (
	"context"
	"flag"
	"fmt"
	stream_chan2 "github.com/snowlyg/go-rtsp-server/stream_chan"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/snowlyg/go-rtsp-server/extend/db"

	figure "github.com/common-nighthawk/go-figure"
	"github.com/go-cmd/cmd"
	"github.com/kardianos/service"
	"github.com/snowlyg/go-rtsp-server/extend/utils"
	"github.com/snowlyg/go-rtsp-server/models"
	"github.com/snowlyg/go-rtsp-server/routers"
)

var (
	gitCommitCode string
	buildDateTime string
)

type program struct {
	httpPort   int
	httpServer *http.Server
}

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
		log.Println("http server end")
	}()
	return
}

func (p *program) Start(s service.Service) (err error) {

	log.Println("********** START **********")
	if utils.IsPortInUse(p.httpPort) {
		err = fmt.Errorf("HTTP port[%d] In Use", p.httpPort)
		return
	}

	err = models.Init()
	if err != nil {
		return
	}
	err = routers.Init()
	if err != nil {
		return
	}

	_ = p.StartHTTP()

	if !utils.Debug {
		log.Println("log files -->", utils.LogDir())
		log.SetOutput(utils.GetLogWriter())
	}

	go func() {
		for range routers.API.RestartChan {
			_ = p.StopHTTP()
			utils.ReloadConf()
			_ = p.StartHTTP()
		}
	}()

	ffmpeg := utils.Conf().Section("rtsp").Key("ffmpeg_path").MustString("ffmpeg")
	stream_chan := stream_chan2.GetStreamChan()
	go func() {
		log.Println("log files -->", 2222)
		pusher2ffmpegMap := make(map[*models.Stream]*cmd.Cmd)
		var stream *models.Stream
		addChnOk := true
		removeChnOk := true
		for addChnOk || removeChnOk {
			select {
			case stream, addChnOk = <-stream_chan.AddPusherCh:
				log.Println("addChnOk -->", stream, addChnOk)
				if addChnOk {
					if stream.Status {
						//	ffmpeg -i rtsp://localhost:8554/original -c:v libx264 -preset ultrafast -tune zerolatency -b 600k -f rtsp rtsp://localhost:8554/compressed

						url := stream.GetUrl()
						s2 := "rtsp"
						if strings.Contains(url, "rtmp://") {
							s2 = "flv"
						}

						params := []string{"-re", "-i", stream.URL, "-strict", "-2", "-vcodec", "libx264", "-max_delay", "100", "-acodec", "aac", "-f", "-g", "5", "-b", "700000", s2, url}
						findCmd := cmd.NewCmd(ffmpeg, params...)
						statusChan := findCmd.Start() // non-blocking
						finalStatus := <-statusChan

						ticker := time.NewTicker(2 * time.Second)

						//Print last line of stdout every 2s
						go func() {
							for range ticker.C {
								// Print each line of STDOUT from Cmd
								for _, line := range finalStatus.Stdout {
									fmt.Println(line)
								}
							}
						}()

						pusher2ffmpegMap[stream] = findCmd
						log.Printf("add ffmpeg [%v] to pull stream from pusher[%v]", findCmd, statusChan)
					}
				} else {
					log.Printf("addPusherChan closed")
				}
			case stream, removeChnOk = <-stream_chan.RemovePusherCh:
				if removeChnOk {
					if !stream.Status {
						findCmd := pusher2ffmpegMap[stream]
						if findCmd != nil {
							findCmd.Stop()
						}
						delete(pusher2ffmpegMap, stream)
						log.Printf("delete ffmpeg from pull stream from pusher[%v]", stream)
					}
				} else {
					for _, findCmd := range pusher2ffmpegMap {
						findCmd.Stop()
					}
					pusher2ffmpegMap = make(map[*models.Stream]*cmd.Cmd)
					log.Printf("removePusherChan closed")
				}
			}
		}

		log.Printf("demon pull streams")
		for {
			var streams []models.Stream
			if err := db.SQLite.Find(&streams).Error; err != nil {
				log.Printf("find stream err:%v", err)
				return
			}

			for i := len(streams) - 1; i > -1; i-- {
				//v := streams[i]
				agent := fmt.Sprintf("go-rtsp-serverGo/%s", routers.BuildVersion)
				if routers.BuildDateTime != "" {
					agent = fmt.Sprintf("%s(%s)", agent, routers.BuildDateTime)
				}
			}
			time.Sleep(2 * time.Second)
		}
	}()
	return
}

func (p *program) Stop(s service.Service) (err error) {
	defer log.Println("********** STOP **********")
	defer utils.CloseLogWriter()
	p.StopHTTP()
	models.Close()
	return
}

func main() {

	flag.StringVar(&utils.FlagVarConfFile, "config", "", "configure file path")
	flag.Parse()
	tail := flag.Args()

	// log
	log.SetPrefix("[go-rtsp-server] ")
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	log.Printf("git commit code:%s", gitCommitCode)
	log.Printf("build date:%s", buildDateTime)

	routers.BuildVersion = fmt.Sprintf("%s.%s", routers.BuildVersion, gitCommitCode)
	routers.BuildDateTime = buildDateTime

	sec := utils.Conf().Section("service")
	svcConfig := &service.Config{
		Name:        sec.Key("name").MustString("go-rtsp-server_Service"),
		DisplayName: sec.Key("display_name").MustString("go-rtsp-server_Service"),
		Description: sec.Key("description").MustString("go-rtsp-server_Service"),
	}

	httpPort := utils.Conf().Section("http").Key("port").MustInt(10008)
	p := &program{
		httpPort: httpPort,
	}
	s, err := service.New(p, svcConfig)
	if err != nil {
		log.Println(err)
		utils.PauseExit()
	}

	if len(tail) > 0 {
		cmd := strings.ToLower(tail[0])
		if cmd == "install" || cmd == "stop" || cmd == "start" || cmd == "uninstall" {
			figure.NewFigure("go-rtsp-server", "", false).Print()
			log.Println(svcConfig.Name, cmd, "...")
			if err = service.Control(s, cmd); err != nil {
				log.Println(err)
				utils.PauseExit()
			}
			log.Println(svcConfig.Name, cmd, "ok")
			return
		}
	}
	figure.NewFigure("go-rtsp-server", "", false).Print()
	if err = s.Run(); err != nil {
		log.Println(err)
		utils.PauseExit()
	}
}
