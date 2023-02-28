package taproot

import (
	"fmt"
	"github.com/google/deck"
	"net/http"
)

func (srv *Server) HandlePanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// can't recover, so fail gracefully and close the connection
				deck.Error("Catching panic() on " + r.URL.String())
				w.Header().Set("Connection", "close")
				srv.ErrorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
