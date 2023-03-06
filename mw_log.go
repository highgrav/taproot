package taproot

import (
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/google/deck"
	"net/http"
)

func (srv *AppServer) HandleLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrId := ""

		if r.Context().Value(CONTEXT_CORRELATION_KEY_NAME) != nil {
			corrId = r.Context().Value(CONTEXT_CORRELATION_KEY_NAME).(string)
		}

		deck.Info(fmt.Sprintf("CALL\t%s\t%s\t\n", corrId, r.URL))

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		// Everything below here will be executed on the way back up the chain
		deck.Info(fmt.Sprintf("RET\t%s\t%d\t\n", corrId, metrics.Code))
	})
}
