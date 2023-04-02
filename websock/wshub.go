package websock

import (
	"context"
	"github.com/highgrav/taproot/v1/common"
	"github.com/highgrav/taproot/v1/logging"
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

type WSHub struct {
	Name  string
	conns map[string][]*WSConn
	acts  chan func()
}

func NewWSHub(id string) *WSHub {
	hub := &WSHub{
		Name:  id,
		conns: make(map[string][]*WSConn),
		acts:  make(chan func()),
	}
	go hub.runInternalActions()
	return hub
}

func (hub *WSHub) AddClient(wsconn *WSConn) {
	hub.acts <- func() {
		if _, ok := hub.conns[wsconn.Key]; !ok {
			hub.conns[wsconn.Key] = []*WSConn{wsconn}
		}
		hub.conns[wsconn.Key] = append(hub.conns[wsconn.Key], wsconn)
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
		if ok {
			wss := make([]*WSConn, 0)
			for _, val := range vals {
				if val != wsconn {
					wss = append(wss, val)
				} else {
					logging.LogToDeck(context.Background(), "info", "WS", "info", "Closing WS conn "+wsconn.Key)
					close(val.Reader)
					close(val.Writer)
					if val.Conn != nil {
						val.Conn.Close()
					}
				}
			}
			hub.conns[wsconn.Key] = wss
		}
	}
}

func (hub *WSHub) RemoveClients(clientId string) {
	hub.acts <- func() {
		if vals, ok := hub.conns[clientId]; ok {
			for _, val := range vals {
				val.closeChan <- true
			}
			delete(hub.conns, clientId)
		}
	}
}

func (hub *WSHub) GenerateNewId(len int) string {
	id := common.CreateRandString(len)
	_, ok := hub.conns[id]
	for ok {
		id = common.CreateRandString(len)
		_, ok = hub.conns[id]
	}
	hub.conns[id] = []*WSConn{}
	return id
}

// TODO -- need to remove expired WSConns
