package websock

import (
	"highgrav/taproot/v1/authn"
	"net"
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
	conns map[string][]WSConn
	acts  chan func()
}

func (hub *WSHub) AddClient(clientId string, conn WSConn) {
	hub.acts <- func() {
		if _, ok := hub.conns[clientId]; !ok {
			hub.conns[clientId] = make([]WSConn, 0)
		}
		hub.conns[clientId] = append(hub.conns[clientId], conn)
	}
}

// send the close channel
func (hub *WSHub) RemoveClient(clientId string, client WSConn) {

}

type WSConn struct {
	Key    string
	User   authn.User
	Conn   net.Conn
	Close  chan bool
	Reader chan []byte
	Writer chan []byte
}
