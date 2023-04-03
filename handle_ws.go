package taproot

import (
	"github.com/gobwas/ws"
	"github.com/highgrav/taproot/v1/authn"
	"github.com/highgrav/taproot/v1/constants"
	"github.com/highgrav/taproot/v1/logging"
	"github.com/highgrav/taproot/v1/websock"
	"net/http"
)

type GenerateWSHandler func() websock.IWebSocketHandler

/*
HandleWS() is a simple handler for creating and running WS connections. Unlike SSEs, you may want to create your own
handler. This could be considered a starting point for a more tailored approach.
*/
func (srv *AppServer) HandleWS(brokerName string, createHandler GenerateWSHandler) http.HandlerFunc {
	if _, ok := srv.WSHubs[brokerName]; !ok {
		panic("Attempted to register websocket handler with non-existent broker name " + brokerName)
	}
	hub, _ := srv.WSHubs[brokerName]
	return func(w http.ResponseWriter, r *http.Request) {
		sessid, ok := r.Context().Value(constants.HTTP_CONTEXT_SESSION_KEY).(string)
		if !ok {
			sessid = ""
		}

		handler := createHandler()
		defer handler.Cancel()
		wsReaderChan, wsWriterChan, err := handler.Init(w, r)
		if err != nil {
			logging.LogToDeck(r.Context(), "error", "WS", "error", "error calling Init() on WS Handler: "+err.Error())
			return
		}

		// We use the zero-copy upgrade approach so we can inject headers (and so on) that might be injected by the
		// WS handler created previously
		headerList := http.Header{}
		for k, v := range w.Header() {
			headerList[k] = v
		}
		upgrader := ws.HTTPUpgrader{
			Header: headerList,
		}
		conn, rw, _, err := upgrader.Upgrade(r, w)

		if err != nil {
			logging.LogToDeck(r.Context(), "error", "ws", "error", "could not upgrade ws conn: "+err.Error())
			srv.ErrorResponse(w, r, 500, "ws server error")
			return
		}

		u, err := authn.GetUserFromRequest(r)
		if sessid == "" && u.UserID == "" {
			sessid = hub.GenerateNewId(16)
		}
		wsc := websock.NewWSConn(sessid, u, conn, rw)
		srv.WSHubs[brokerName].AddClient(&wsc)
		logging.LogToDeck(r.Context(), "info", "WS", "info", "opening WS handler")
		defer srv.WSHubs[brokerName].RemoveClient(&wsc)

		for {
			select {
			case isDone := <-wsc.CloseChan:
				if isDone {
					logging.LogToDeck(r.Context(), "info", "WS", "info", "closing WS handler")
					return
				}
			case inc := <-wsc.Reader:
				wsReaderChan <- inc
			case outg := <-wsWriterChan:
				wsc.Writer <- outg
			}
		}
	}
}
