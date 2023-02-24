package taproot

import (
	"fmt"
	"net/http"
)

func (srv *Server) HandlePanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// can't recover, so fail gracefully and close the connection
				w.Header().Set("Connection", "close")
				srv.ErrorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
