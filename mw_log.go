package taproot

import (
	"github.com/felixge/httpsnoop"
	"github.com/tomasen/realip"
	"net/http"
	"time"
)

func (srv *AppServer) HandleLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqTime := time.Now()
		corrId := ""
		clientIp := realip.FromRequest(r)
		if r.Context().Value(HTTP_CONTEXT_CORRELATION_KEY) != nil {
			corrId = r.Context().Value(HTTP_CONTEXT_CORRELATION_KEY).(string)
		}

		//		deck.Info(fmt.Sprintf("REQ\t%s\t%s\t-\t%s\t%s\t%s\t\t\t\n", clientIp, corrId, reqTime.Format(customTimeFormat), r.Method, r.URL))
		LogW3CRequest("info", reqTime, clientIp, corrId, r.Method, r.URL.String())

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		// Everything below here will be executed on the way back up the chain

		//	deck.Info(fmt.Sprintf("RES\t%s\t%s\t-\t%s\t%s\t%s\t%s\t%d\t%d\n", clientIp, corrId, time.Now().Format(customTimeFormat), time.Now().Sub(reqTime).String(), r.Method, r.URL, metrics.Code, metrics.Written))
		LogW3CResponse("info", reqTime, clientIp, corrId, r.Method, r.URL.String(), metrics.Code, int(metrics.Written))
	})
}
