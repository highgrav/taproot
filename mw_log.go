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

		logging.LogW3CRequest("info", reqTime, clientIp, r.Context(), r.Method, r.URL.String(), userId)

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		// Everything below here will be executed on the way back up the chain
		logging.LogW3CResponse("info", reqTime, clientIp, r.Context(), r.Method, r.URL.String(), metrics.Code, int(metrics.Written), userId)
	})
}
