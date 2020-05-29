package rtsp

import (
	"fmt"
	"github.com/snowlyg/GoEasyFfmpeg/extend/EasyGoLib/utils"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Server struct {
	SessionLogger
	//TCPListener    *net.TCPListener
	//TCPPort        int
	Stoped         bool
	pushers        map[string]*Pusher // Path <-> Pusher
	pushersLock    sync.RWMutex
	addPusherCh    chan *Pusher
	removePusherCh chan *Pusher
}

var Instance *Server = &Server{
	SessionLogger:  SessionLogger{log.New(os.Stdout, "[RTSPServer]", log.LstdFlags|log.Lshortfile)},
	Stoped:         true,
	pushers:        make(map[string]*Pusher),
	addPusherCh:    make(chan *Pusher),
	removePusherCh: make(chan *Pusher),
}

func GetServer() *Server {
	return Instance
}

// Start 启动
func (server *Server) Start() (err error) {
	logger := server.logger
	ffmpeg := utils.Conf().Section("rtsp").Key("ffmpeg_path").MustString("")
	m3u8DirPath := utils.Conf().Section("rtsp").Key("m3u8_dir_path").MustString("")

	go func() { // 保持到本地
		pusher2FfmpegMap := make(map[*Pusher]*exec.Cmd)
		logger.Printf("Prepare to save stream to local....")
		defer logger.Printf("End save stream to local....")
		var pusher *Pusher
		addChnOk := true
		removeChnOk := true
		for addChnOk || removeChnOk {
			select {
			case pusher, addChnOk = <-server.addPusherCh:
				if addChnOk {
					dir := path.Join(m3u8DirPath, "logs", time.Now().Format("20060102"))
					err := utils.EnsureDir(dir)
					if err != nil {
						logger.Printf("EnsureDir:[%s] err:%v.", dir, err)
						continue
					}

					paramStr := utils.Conf().Section("rtsp").Key("decoder").MustString("-strict -2 -threads 2 -c:v copy -c:a copy -f rtsp")
					paramsOfThisPath := strings.Split(paramStr, " ")
					params := []string{"-i", pusher.Source, pusher.Path}
					params = append(params[:2], append(paramsOfThisPath, params[2:]...)...)

					cmd := exec.Command(ffmpeg, params...)
					f, err := os.OpenFile(path.Join(dir, fmt.Sprintf("log.txt")), os.O_RDWR|os.O_CREATE, 0755)
					if err == nil {
						cmd.Stdout = f
						cmd.Stderr = f
					}
					err = cmd.Start()
					if err != nil {
						logger.Printf("Start ffmpeg err:%v", err)
					}

					pusher2FfmpegMap[pusher] = cmd
					logger.Printf("add ffmpeg [%v] to pull stream from pusher[%v]", cmd, pusher)

				} else {
					logger.Printf("addPusherChan closed")
				}
			case pusher, removeChnOk = <-server.removePusherCh:
				if removeChnOk {
					cmd := pusher2FfmpegMap[pusher]
					proc := cmd.Process
					if proc != nil {
						logger.Printf("prepare to SIGTERM to process:%v", proc)
						proc.Signal(syscall.SIGTERM)
						proc.Wait()

						logger.Printf("process:%v terminate.", proc)
					}
					delete(pusher2FfmpegMap, pusher)
					logger.Printf("delete ffmpeg from pull stream from pusher[%v]", pusher)
				} else {
					for _, cmd := range pusher2FfmpegMap {
						proc := cmd.Process
						if proc != nil {
							logger.Printf("prepare to SIGTERM to process:%v", proc)
							proc.Signal(syscall.SIGTERM)
						}
					}
					pusher2FfmpegMap = make(map[*Pusher]*exec.Cmd)
					logger.Printf("removePusherChan closed")
				}
			}
		}
	}()

	server.Stoped = false

	return
}

// Stop 停止
func (server *Server) Stop() {
	server.Stoped = true
	server.pushersLock.Lock()
	server.pushers = make(map[string]*Pusher)
	server.pushersLock.Unlock()

	close(server.addPusherCh)
	close(server.removePusherCh)
}

// AddPusher 添加推流进程
func (server *Server) AddPusher(pusher *Pusher) bool {
	logger := server.logger
	added := false
	server.pushersLock.Lock()
	_, ok := server.pushers[pusher.Path]
	if !ok {
		server.pushers[pusher.Path] = pusher
		logger.Printf("%v start, now pusher size[%d]", pusher, len(server.pushers))
		added = true
	} else {
		added = false
	}
	server.pushersLock.Unlock()
	if added {
		go pusher.Start()
		server.addPusherCh <- pusher
	}

	return added
}

// RemovePusher 移除推流
func (server *Server) RemovePusher(pusher *Pusher) {
	logger := server.logger
	removed := false
	server.pushersLock.Lock()
	if _pusher, ok := server.pushers[pusher.Path]; ok && pusher.ID == _pusher.ID {
		delete(server.pushers, pusher.Path)
		logger.Printf("%v end, now pusher size[%d]\n", pusher, len(server.pushers))
		removed = true
	}
	server.pushersLock.Unlock()
	if removed {
		server.removePusherCh <- pusher
	}
}

// GetPusher 获取推流
func (server *Server) GetPusher(path string) (pusher *Pusher) {
	server.pushersLock.RLock()
	pusher = server.pushers[path]
	server.pushersLock.RUnlock()
	return
}

// GetPushers 获取推流列表
func (server *Server) GetPushers() (pushers map[string]*Pusher) {
	pushers = make(map[string]*Pusher)
	server.pushersLock.RLock()
	for k, v := range server.pushers {
		pushers[k] = v
	}
	server.pushersLock.RUnlock()
	return
}

// GetPusherSize 获取推流数量
func (server *Server) GetPusherSize() (size int) {
	server.pushersLock.RLock()
	size = len(server.pushers)
	server.pushersLock.RUnlock()
	return
}
