package taproot

import (
	"highgrav/taproot/v1/authn"
	"net/http"
)

func (srv *AppServer) HandleUserInjection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO
		next.ServeHTTP(w, r)
	})
}

func (srv *AppServer) GetUserFromRequest(r *http.Request) authn.User {
	// TODO
	return authn.User{}
}
