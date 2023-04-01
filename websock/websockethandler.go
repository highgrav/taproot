package websock

import "net/http"

type WebSocketHandler func(r *http.Request, wsconn WSConn, autoTimeoutMinutes int, args ...any)
