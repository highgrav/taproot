package websock

import (
	"bufio"
	"context"
	"github.com/gobwas/ws/wsutil"
	"github.com/highgrav/taproot/v1/authn"
	"github.com/highgrav/taproot/v1/logging"
	"net"
)

type WSConn struct {
	Key       string
	User      authn.User
	Conn      net.Conn
	Buf       *bufio.ReadWriter
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
		CloseChan: make(chan bool),
		Reader:    make(chan WSFrame),
		Writer:    make(chan WSFrame),
	}
	go wsc.process()
	return wsc
}

func (wsc *WSConn) Close() {
	if wsc.CloseChan != nil {
		wsc.CloseChan <- true
	}
}

func (wsc *WSConn) process() {

	var isClosed bool = false

	// write
	go func() {
		for {
			select {
			case toWrite := <-wsc.Writer:
				if isClosed {
					return
				}
				err := wsutil.WriteServerMessage(wsc.Conn, toWrite.Op, toWrite.Data)
				if err != nil {
					wsc.Close()
					logging.LogToDeck(context.Background(), "error", "WS", "error", "caught error writing ws client data in "+wsc.Key+": "+err.Error())
					return
				}
			}
		}
		wsc.Conn.Close()
	}()

	// read
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
		wsc.Conn.Close()
	}()

	// wait for a close
	go func() {
		for {
			select {
			case closed := <-wsc.CloseChan:
				if closed {
					isClosed = true
					close(wsc.Reader)
					close(wsc.Writer)
					wsc.Conn.Close()
					return
				}
			}
		}
		wsc.Conn.Close()
	}()

}
