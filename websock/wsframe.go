package websock

import "github.com/gobwas/ws"

type WSFrame struct {
	Op   ws.OpCode
	Data []byte ``
}
