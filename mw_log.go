package taproot

import (
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/google/deck"
	"github.com/tomasen/realip"
	"net/http"
	"time"
)

func (srv *AppServer) HandleLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqTime := time.Now()
		corrId := ""
		clientIp := realip.FromRequest(r)
		customTimeFormat := "2006-01-02T15:04:05.000-07:00"
		if r.Context().Value(CONTEXT_CORRELATION_KEY_NAME) != nil {
			corrId = r.Context().Value(CONTEXT_CORRELATION_KEY_NAME).(string)
		}

		deck.Info(fmt.Sprintf("REQ\t%s\t%s\t-\t%s\t%s\t%s\t\t\t\n", clientIp, corrId, reqTime.Format(customTimeFormat), r.Method, r.URL))

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		// Everything below here will be executed on the way back up the chain

		deck.Info(fmt.Sprintf("RES\t%s\t%s\t-\t%s\t%s\t%s\t%s\t%d\t%d\n", clientIp, corrId, time.Now().Format(customTimeFormat), time.Now().Sub(reqTime).String(), r.Method, r.URL, metrics.Code, metrics.Written))
	})
}
