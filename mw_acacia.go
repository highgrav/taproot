package taproot

import (
	"net/http"
)

// This assumes that we've already injected a user into the request
func (srv *Server) HandleRightsInjection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
