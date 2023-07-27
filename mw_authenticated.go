package taproot

import (
	"github.com/highgrav/taproot/authn"
	"net/http"
)

func (srv *AppServer) HandleAuthOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := authn.GetUserFromRequest(r)
		if err != nil || u.UserID == "" {
			srv.ErrorResponse(w, r, http.StatusUnauthorized, "not authorized")
		}
		next.ServeHTTP(w, r)
	})
}
