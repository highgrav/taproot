package taproot

import (
	"github.com/google/deck"
	"github.com/tomasen/realip"
	"net/http"
)

func (srv *AppServer) handleIPFiltering(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rip := realip.FromRequest(r)
		if !srv.httpIpFilter.Allowed(rip) {
			deck.Error("IP filter blocked IP " + rip)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
