package taproot

import (
	"expvar"
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func (srv *AppServer) HandleMetrics(next http.Handler) http.Handler {
	type stats struct {
		requests       *expvar.Int
		responses      *expvar.Int
		processingTime *expvar.Int
		responseCodes  *expvar.Map
	}
	var globalStats stats = stats{
		requests:       expvar.NewInt("total requests received"),
		responses:      expvar.NewInt("total responses sent"),
		processingTime: expvar.NewInt("total processing time in microsecs"),
		responseCodes:  expvar.NewMap("total responses by HTTP code"),
	}
	var routeStates = make(map[string]stats)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ps := httprouter.ParamsFromContext(r.Context())
		registeredPath := ps.MatchedRoutePath()
		fmt.Println("PATH: " + registeredPath)
		stat, exists := routeStates[registeredPath]
		if !exists {
			// There's a minor race condition here, but it's not important
			routeStates[registeredPath] = stats{
				requests:       expvar.NewInt(r.URL.Path + ": requests received"),
				responses:      expvar.NewInt(r.URL.Path + ": responses sent"),
				processingTime: expvar.NewInt(r.URL.Path + ": processing time in microsecs"),
				responseCodes:  expvar.NewMap(r.URL.Path + ": responses by HTTP code"),
			}
			stat, _ = routeStates[registeredPath]
		}
		globalStats.requests.Add(1)
		stat.requests.Add(1)

		// keep going
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		// pick up processing on the way back
		stat, _ = routeStates[registeredPath]
		globalStats.responses.Add(1)
		stat.responses.Add(1)
		globalStats.processingTime.Add(metrics.Duration.Microseconds())
		stat.processingTime.Add(metrics.Duration.Microseconds())
		c := strconv.Itoa(metrics.Code)
		globalStats.responseCodes.Add(c, 1)
		stat.responseCodes.Add(c, 1)
	})
}
