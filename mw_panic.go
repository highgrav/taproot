package taproot

import (
	"fmt"
	"github.com/highgrav/taproot/logging"
	"net/http"
	"runtime/debug"
)

func (srv *AppServer) HandlePanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// can't recover, so fail gracefully and close the connection
				logging.LogToDeck(r.Context(), "fatal", "TAPROOT", "panic", "catching panic() on "+r.URL.String())
				logging.LogToDeck(r.Context(), "fatal", "TAPROOT", "panic", err.(error).Error())
				panicTrace := string(debug.Stack())
				fn, err := srv.DumpStackTrace(panicTrace)
				if err != nil {
					logging.LogToDeck(r.Context(), "fatal", "TAPROOT", "panic", "Failed to write stack trace: "+err.Error())
				} else {
					logging.LogToDeck(r.Context(), "fatal", "TAPROOT", "panic", "Wrote stack trace to "+fn)
				}
				w.Header().Set("Connection", "close")
				srv.ErrorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
