package stream_chan

import (
	"github.com/snowlyg/go-rtsp-server/models"
)

type streamChan struct {
	AddPusherCh    chan *models.Stream
	RemovePusherCh chan *models.Stream
}

var Instance *streamChan = &streamChan{
	AddPusherCh:    make(chan *models.Stream),
	RemovePusherCh: make(chan *models.Stream),
}

func GetStreamChan() *streamChan {
	return Instance
}
