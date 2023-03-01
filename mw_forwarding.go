package taproot

import (
	"net/http"
	"strings"
)

// TODO --tomasen/realip is probably a better option here
func (srv *Server) HandleForwarding(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ip string = r.RemoteAddr
		xff := r.Header.Get("X-Forwarded-For")
		if xff != "" {
			ip = xff
			ips := strings.Split(ip, ", ")
			if len(ips) > 1 {
				ip = ips[0]
			}
		}
		next.ServeHTTP(w, r)
	})
}
