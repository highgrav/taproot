package taproot

import (
	"github.com/felixge/httpsnoop"
	"github.com/highgrav/taproot/v1/authn"
	"github.com/highgrav/taproot/v1/constants"
	"github.com/highgrav/taproot/v1/logging"
	"github.com/tomasen/realip"
	"net/http"
	"time"
)

func (srv *AppServer) HandleLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqTime := time.Now()
		clientIp := realip.FromRequest(r)
		user, ok := r.Context().Value(constants.HTTP_CONTEXT_USER_KEY).(authn.User)
		var userId string = "-"
		if ok && user.UserID != "" {
			userId = user.UserID
		}

		//		deck.Info(fmt.Sprintf("REQ\t%s\t%s\t-\t%s\t%s\t%s\t\t\t\n", clientIp, corrId, reqTime.Format(customTimeFormat), r.Method, r.URL))
		logging.LogW3CRequest("info", reqTime, clientIp, r.Context(), r.Method, r.URL.String(), userId)

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		// Everything below here will be executed on the way back up the chain

		//	deck.Info(fmt.Sprintf("RES\t%s\t%s\t-\t%s\t%s\t%s\t%s\t%d\t%d\n", clientIp, corrId, time.Now().Format(customTimeFormat), time.Now().Sub(reqTime).String(), r.Method, r.URL, metrics.Code, metrics.Written))
		logging.LogW3CResponse("info", reqTime, clientIp, r.Context(), r.Method, r.URL.String(), metrics.Code, int(metrics.Written), userId)
	})
}
