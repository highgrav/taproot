package taproot

import (
	"highgrav/taproot/v1/sse"
	"net/http"
	"time"
)

// This is a generic middleware for connecting to a broker and handling messages.
// Often you'll need specific logic, so you'd want to write your own handler,
// but this is a decent starting point. Set autoTimeoutMinutes to something
// reasonably far in the future -- 72 hours or so.
func (srv *AppServer) HandleSSE(brokerName string, autoTimeoutMinutes int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if autoTimeoutMinutes < 1 {
			autoTimeoutMinutes = 525600
		}
		timer := time.NewTimer(time.Duration(autoTimeoutMinutes) * time.Minute)
		broker, ok := srv.SseBrokers[brokerName]
		user := srv.GetUserFromRequest(r)
		if !ok {
			// return error
			srv.ErrorResponse(w, r, 500, "message source not available")
			return
		}
		fl := w.(http.Flusher)
		w.Header().Set("Content-Type", sse.SSE_MIMETYPE)
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := make(chan sse.SSEEvent)
		broker.AddClient(user.UserID, ch)
		defer broker.RemoveClient(user.UserID, ch)
		for {
			select {
			case <-timer.C:
				// auto-timeout to deal with absurdly long-lived connections
				return
			case <-r.Context().Done():
				// client's broken the connection
				return
			case msg := <-ch:
				w.Write([]byte(msg.Dispatch()))
				fl.Flush()
			}
		}
	}
}
