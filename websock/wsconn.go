package websock

import (
	"bufio"
	"context"
	"github.com/gobwas/ws/wsutil"
	"github.com/highgrav/taproot/authn"
	"github.com/highgrav/taproot/logging"
	"net"
)

type WSConn struct {
	Key       string
	User      authn.User
	Conn      net.Conn
	Buf       *bufio.ReadWriter
	closeChan chan bool
	CloseChan chan bool
	Reader    chan WSFrame
	Writer    chan WSFrame
}

func NewWSConn(id string, user authn.User, conn net.Conn, buf *bufio.ReadWriter) WSConn {
	wsc := WSConn{
		Key:       id,
		User:      user,
		Conn:      conn,
		Buf:       buf,
		closeChan: make(chan bool),
		CloseChan: make(chan bool),
		Reader:    make(chan WSFrame),
		Writer:    make(chan WSFrame),
	}
	go wsc.process()
	return wsc
}

func (wsc *WSConn) Close() {
	if wsc.closeChan != nil {
		wsc.closeChan <- true
	}
}

func (wsc *WSConn) process() {
	var isClosed bool = false
	// write
	go func() {
		for {
			select {
			case closed := <-wsc.closeChan:
				if closed {
					isClosed = true
					wsc.CloseChan <- true
					return
				}
			case toWrite := <-wsc.Writer:
				if isClosed {
					return
				}
				err := wsutil.WriteServerMessage(wsc.Conn, toWrite.Op, toWrite.Data)
				if err != nil {
					logging.LogToDeck(context.Background(), "error", "WS", "error", "caught error writing ws client data in "+wsc.Key+": "+err.Error())
					wsc.Close()
					return
				}
			}
		}
	}()

	// read -- this should break as soon as the connection fails, so we don't need to clean up
	go func() {
		for {
			msg, op, err := wsutil.ReadClientData(wsc.Conn)
			if err != nil {
				logging.LogToDeck(context.Background(), "error", "WS", "error", "caught error reading ws client data in "+wsc.Key+": "+err.Error())
				wsc.Close()
				return
			}
			if isClosed {
				return
			}
			wsc.Reader <- WSFrame{
				Op:   op,
				Data: msg,
			}
		}
	}()

}
