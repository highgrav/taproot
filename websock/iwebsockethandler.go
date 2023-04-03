package websock

import "net/http"

type IWebSocketHandler interface {
	//	Init(r *http.Request, wsconn WSConn, autoTimeoutMinutes int, args ...any)
	Init(w http.ResponseWriter, r *http.Request) (wsReader, wsWriter chan WSFrame, err error)
	GetChannels() (wsReader, wsWriter chan WSFrame, err error)
	Cancel() error
}
