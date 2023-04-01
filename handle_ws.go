package taproot

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/highgrav/taproot/v1/authn"
	"github.com/highgrav/taproot/v1/constants"
	"github.com/highgrav/taproot/v1/logging"
	"github.com/highgrav/taproot/v1/websock"
	"net/http"
)

/*
HandleWS() is a simple handler for creating and running WS connections. Unlike SSEs, you probably want to create your own
handler.
*/
func (srv *AppServer) HandleWS(brokerName string, wsHandler websock.WebSocketHandler, autoTimeoutMinutes int) http.HandlerFunc {
	fmt.Println(">>>>>>>WEBSOCKET CALL")
	if _, ok := srv.WSHubs[brokerName]; !ok {
		panic("Attempted to register websocket handler with non-existent broker name " + brokerName)
	}
	hub, _ := srv.WSHubs[brokerName]
	return func(w http.ResponseWriter, r *http.Request) {
		sessid, ok := r.Context().Value(constants.HTTP_CONTEXT_SESSION_KEY).(string)
		if !ok {
			sessid = ""
		}
		if autoTimeoutMinutes < 1 {
			autoTimeoutMinutes = 525600
		}
		conn, rw, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			logging.LogToDeck(r.Context(), "error", "ws", "error", "could not upgrade ws conn: "+err.Error())
			srv.ErrorResponse(w, r, 500, "ws server error")
			return
		}

		u, err := authn.GetUserFromRequest(r)
		if sessid == "" && u.UserID == "" {
			// TODO -- not a great idea, need to revisit?
			sessid = hub.GenerateNewId(16)
		}
		wsc := websock.NewWSConn(sessid, u, conn, rw)
		srv.WSHubs[brokerName].AddClient(wsc)
		logging.LogToDeck(r.Context(), "info", "WS", "info", "opening WS handler")

		wsHandler(r, wsc, autoTimeoutMinutes)
		logging.LogToDeck(r.Context(), "info", "WS", "info", "closing WS handler")
	}
}
