package rtsp

import (
	"github.com/snowlyg/GoEasyFfmpeg/extend/utils"
	"sync"
)

type Player struct {
	Pusher               *Pusher
	cond                 *sync.Cond
	queueLimit           int
	dropPacketWhenPaused bool
	paused               bool
	Stoped               bool
}

// NewPlayer 新建播放器
func NewPlayer(pusher *Pusher) (player *Player) {
	queueLimit := utils.Conf().Section("rtsp").Key("player_queue_limit").MustInt(0)
	dropPacketWhenPaused := utils.Conf().Section("rtsp").Key("drop_packet_when_paused").MustInt(0)
	player = &Player{
		//Session:              session,
		Pusher: pusher,
		cond:   sync.NewCond(&sync.Mutex{}),
		//queue:                make([]*RTPPack, 0),
		queueLimit:           queueLimit,
		dropPacketWhenPaused: dropPacketWhenPaused != 0,
		paused:               false,
	}
	//session.StopHandles = append(session.StopHandles, func() {
	//	pusher.RemovePlayer(player)
	//	player.cond.Broadcast()
	//})
	return
}

// Start 启动播放器
func (player *Player) Start() {
	for !player.Stoped {
	}
}

// Pause 暂停播放器
func (player *Player) Pause(paused bool) {

	player.paused = paused
	player.cond.L.Unlock()
}
