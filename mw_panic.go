package taproot

import (
	"fmt"
	"highgrav/taproot/v1/logging"
	"net/http"
)

func (srv *AppServer) HandlePanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// can't recover, so fail gracefully and close the connection
				logging.LogToDeck("fatal", "PANIC\tCatching panic() on "+r.URL.String())
				logging.LogToDeck("fatal", err.(error).Error())
				w.Header().Set("Connection", "close")
				srv.ErrorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
