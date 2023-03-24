package taproot

import (
	"github.com/highgrav/taproot/v1/logging"
	"github.com/tomasen/realip"
	"net/http"
)

func (srv *AppServer) handleIPFiltering(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rip := realip.FromRequest(r)
		if !srv.httpIpFilter.Allowed(rip) {
			logging.LogToDeck(r.Context(), "warn", "IPFILTER", "alert", "IP filter blocked IP "+rip)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
