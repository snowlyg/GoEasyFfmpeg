package rtsp

import (
	"github.com/snowlyg/go-rtsp-server/extend/EasyGoLib/utils"
	"log"
	"os"
	"os/exec"
	"path"
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
	SessionLogger: SessionLogger{log.New(os.Stdout, "[RTSPServer]", log.LstdFlags|log.Lshortfile)},
	Stoped:        true,
	//TCPPort:        utils.Conf().Section("rtsp").Key("port").MustInt(554),
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

	localRecord := utils.Conf().Section("rtsp").Key("save_stream_to_local").MustInt(0)
	ffmpeg := utils.Conf().Section("rtsp").Key("ffmpeg_path").MustString("")
	m3u8DirPath := utils.Conf().Section("rtsp").Key("m3u8_dir_path").MustString("")
	//tsDurationSecond := utils.Conf().Section("rtsp").Key("ts_duration_second").MustInt(6)

	SaveStreamToLocal := false
	if (len(ffmpeg) > 0) && localRecord > 0 && len(m3u8DirPath) > 0 {
		err := utils.EnsureDir(m3u8DirPath)
		if err != nil {
			logger.Printf("Create m3u8_dir_path[%s] err:%v.", m3u8DirPath, err)
		} else {
			SaveStreamToLocal = true
		}
	}

	go func() { // 保持到本地
		pusher2FfmpegMap := make(map[*Pusher]*exec.Cmd)
		if SaveStreamToLocal {
			logger.Printf("Prepare to save stream to local....")
			defer logger.Printf("End save stream to local....")
		}
		var pusher *Pusher
		addChnOk := true
		removeChnOk := true
		for addChnOk || removeChnOk {
			select {
			case pusher, addChnOk = <-server.addPusherCh:
				if SaveStreamToLocal {
					if addChnOk {
						dir := path.Join(m3u8DirPath, pusher.Path, time.Now().Format("20060102"))
						err := utils.EnsureDir(dir)
						if err != nil {
							logger.Printf("EnsureDir:[%s] err:%v.", dir, err)
							continue
						}

						params := []string{"-i", "rtsp://183.59.168.27:554/PLTV/88888905/224/3221227255/10000100000000060000000001066420_0.smil?icip=88888888", "-strict", "-2", "-vcodec", "h264", "-acodec", "aac", "-f", "flv", "rtmp://localhost:1935/live/rfBd56ti2SMtYvSgD5xAV0YU99zampta7Z7S575KLkIZ9PYk"}
						cmd := exec.Command(ffmpeg, params...)
						err = cmd.Start()
						if err != nil {
							logger.Printf("Start ffmpeg err:%v", err)
						}

						if err := cmd.Wait(); err != nil {
							log.Printf("Cmd returned error: %v", err)
						}

						pusher2FfmpegMap[pusher] = cmd
						logger.Printf("add ffmpeg [%v] to pull stream from pusher[%v]", cmd, pusher)
					} else {
						logger.Printf("addPusherChan closed")
					}
				}
			case pusher, removeChnOk = <-server.removePusherCh:
				if SaveStreamToLocal {
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
