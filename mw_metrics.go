package taproot

import (
	"expvar"
	"github.com/felixge/httpsnoop"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

type stats struct {
	requests       *expvar.Int
	responses      *expvar.Int
	processingTime *expvar.Int
	responseCodes  *expvar.Map
}

func (srv *AppServer) handleLocalMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ps := httprouter.ParamsFromContext(r.Context())
		registeredPath := ps.MatchedRoutePath() // TODO -- broken!
		stat, exists := srv.stats[registeredPath]
		if !exists {
			// There's a minor race condition here, but it's not important
			srv.stats[registeredPath] = stats{
				requests:       expvar.NewInt(r.URL.Path + ": requests received"),
				responses:      expvar.NewInt(r.URL.Path + ": responses sent"),
				processingTime: expvar.NewInt(r.URL.Path + ": processing time in microsecs"),
				responseCodes:  expvar.NewMap(r.URL.Path + ": responses by HTTP code"),
			}
			stat, _ = srv.stats[registeredPath]

		}
		stat.requests.Add(1)

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		stat, _ = srv.stats[registeredPath]
		stat.responses.Add(1)
		stat.processingTime.Add(metrics.Duration.Microseconds())
		c := strconv.Itoa(metrics.Code)
		stat.responseCodes.Add(c, 1)
	})
}

func (srv *AppServer) HandleGlobalMetrics(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.globalStats.requests.Add(1)

		// keep going
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		// pick up processing on the way back
		srv.globalStats.responses.Add(1)
		srv.globalStats.processingTime.Add(metrics.Duration.Microseconds())
		c := strconv.Itoa(metrics.Code)
		srv.globalStats.responseCodes.Add(c, 1)

	})
}
