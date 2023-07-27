package websock

import (
	"context"
	"github.com/highgrav/taproot/common"
	"github.com/highgrav/taproot/logging"
	"sync"
	"sync/atomic"
)

// TODO

const (
	WS_HEADER_CLIENT_UPGRADE           string = "upgrade"
	WS_HEADER_CLIENT_CONNECTION        string = "connection"
	WS_HEADER_SEC_WEBSOCKET_KEY        string = "Sec-WebSocket-Key"
	WS_HEADER_SEC_WEBSOCKET_EXTENSIONS string = "Sec-WebSocket-Extensions"
	WS_HEADER_SEC_WESOCKET_ACCEPT      string = "Sec-WebSocket-Accept"
	WS_HEADER_SEC_WEBSOCKET_PROTOCOL   string = "Sec-WebSocket-Protocol"
	WS_HEADER_SEC_WEBSOCKET_VERSION    string = "Sec-WebSocket-Version"
)

type WSConnContainer struct {
	sync.Mutex
	Conns []*WSConn
}

type WSHub struct {
	sync.Mutex
	Name       string
	Metrics    *WSMetrics
	conns      map[string]*WSConnContainer
	acts       chan func()
	TotalConns int32
}

func NewWSHub(id string) *WSHub {
	hub := &WSHub{
		Name:       id,
		Metrics:    &WSMetrics{},
		conns:      make(map[string]*WSConnContainer),
		acts:       make(chan func()),
		TotalConns: 0,
	}

	go hub.runInternalActions()
	return hub
}

func (hub *WSHub) AddClient(wsconn *WSConn) {
	hub.acts <- func() {
		if _, ok := hub.conns[wsconn.Key]; !ok {
			hub.conns[wsconn.Key] = &WSConnContainer{
				Conns: []*WSConn{wsconn},
			}
		}
		hub.conns[wsconn.Key].Lock()
		hub.conns[wsconn.Key].Conns = append(hub.conns[wsconn.Key].Conns, wsconn)
		atomic.AddInt32(&hub.TotalConns, 1)
		hub.conns[wsconn.Key].Unlock()
		logging.LogToDeck(context.Background(), "info", "WS", "info", "Adding WS conn "+wsconn.Key)
	}
}

func (hub *WSHub) runInternalActions() {
	for act := range hub.acts {
		act()
	} // infinite loop
}

func (hub *WSHub) RemoveClient(wsconn *WSConn) {
	hub.acts <- func() {
		if wsconn == nil {
			return
		}
		vals, ok := hub.conns[wsconn.Key]
		if !ok {
			return
		}
		vals.Lock()
		wss := make([]*WSConn, 0)
		for _, val := range vals.Conns {
			if val != wsconn {
				wss = append(wss, val)
			} else {
				logging.LogToDeck(context.Background(), "info", "WS", "info", "Closing WS conn "+wsconn.Key)
				close(val.Reader)
				close(val.Writer)
				if val.Conn != nil {
					val.Conn.Close()
				}
				atomic.AddInt32(&hub.TotalConns, -1)
			}
		}

		hub.conns[wsconn.Key].Conns = wss
		if len(wss) == 0 {
			delete(hub.conns, wsconn.Key)
		}
		vals.Unlock()
	}
}

func (hub *WSHub) RemoveClients(clientId string) {
	hub.acts <- func() {
		if vals, ok := hub.conns[clientId]; ok {
			vals.Lock()
			for _, val := range vals.Conns {
				val.closeChan <- true
				atomic.AddInt32(&hub.TotalConns, -1)
			}
			vals.Unlock()
		}
		delete(hub.conns, clientId)
	}
}

func (hub *WSHub) GenerateNewId(len int) string {
	hub.Lock()
	id := common.CreateRandString(len)
	_, ok := hub.conns[id]
	for ok {
		id = common.CreateRandString(len)
		_, ok = hub.conns[id]
	}
	hub.conns[id] = &WSConnContainer{
		Conns: make([]*WSConn, 0),
	}
	hub.Unlock()
	return id
}
