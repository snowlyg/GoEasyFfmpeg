package rtsp

import (
	"sync"
)

type Pusher struct {
	ID                string
	Path              string
	players           map[string]*Player //SessionID <-> Player
	playersLock       sync.RWMutex
	gopCacheEnable    bool
	gopCacheLock      sync.RWMutex
	spsppsInSTAPaPack bool
	cond              *sync.Cond
	Stoped            bool
}

// NewClientPusher 新建推流器客户端
func NewClientPusher() (pusher *Pusher) {
	pusher = &Pusher{
		players: make(map[string]*Player),
		cond:    sync.NewCond(&sync.Mutex{}),
	}

	return
}

// NewPusher 新建推流器
func NewPusher() (pusher *Pusher) {
	pusher = &Pusher{
		players: make(map[string]*Player),
		cond:    sync.NewCond(&sync.Mutex{}),
	}

	return
}

// Start 启动推流器
func (pusher *Pusher) Start() {
	for !pusher.Stoped {
		pusher.cond.L.Lock()
	}
}

// Stop 停止推流器
func (pusher *Pusher) Stop() {

}

// GetPlayers 获取播放者
func (pusher *Pusher) GetPlayers() (players map[string]*Player) {
	players = make(map[string]*Player)
	pusher.playersLock.RLock()
	for k, v := range pusher.players {
		players[k] = v
	}
	pusher.playersLock.RUnlock()
	return
}

// HasPlayer 是否在播放
func (pusher *Pusher) HasPlayer(player *Player) bool {
	pusher.playersLock.Lock()
	_, ok := pusher.players[player.Pusher.ID]
	pusher.playersLock.Unlock()
	return ok
}

// AddPlayer 添加播放者
func (pusher *Pusher) AddPlayer(player *Player) *Pusher {

	if pusher.gopCacheEnable {
		pusher.gopCacheLock.RLock()

		pusher.gopCacheLock.RUnlock()
	}

	pusher.playersLock.Lock()
	if _, ok := pusher.players[player.Pusher.ID]; !ok {
		pusher.players[player.Pusher.ID] = player
		go player.Start()
	}
	pusher.playersLock.Unlock()
	return pusher
}

// RemovePlayer 移除播放者
func (pusher *Pusher) RemovePlayer(player *Player) *Pusher {

	pusher.playersLock.Lock()
	if len(pusher.players) == 0 {
		pusher.playersLock.Unlock()
		return pusher
	}
	delete(pusher.players, player.Pusher.ID)

	pusher.playersLock.Unlock()
	return pusher
}

// ClearPlayer 清除播放者
func (pusher *Pusher) ClearPlayer() {
	players := make(map[string]*Player)
	pusher.playersLock.Lock()
	for k, v := range pusher.players {
		v.Stoped = true
		players[k] = v
	}
	pusher.players = make(map[string]*Player)
	pusher.playersLock.Unlock()
	go func() { // do not block
		for _, v := range players {
			v.Stoped = true
		}
	}()
}
