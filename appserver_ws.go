package taproot

import (
	"github.com/highgrav/taproot/websock"
)

func (srv *AppServer) AddWSHub(name string) {
	if srv.WSHubs == nil {
		srv.WSHubs = make(map[string]*websock.WSHub)
	}
	if _, ok := srv.WSHubs[name]; ok {
		return
	}
	wsh := websock.NewWSHub(name)
	srv.WSHubs[name] = wsh
}
