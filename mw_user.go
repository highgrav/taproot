package taproot

import (
	"net/http"
)

func (srv *AppServer) HandleUserInjection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
