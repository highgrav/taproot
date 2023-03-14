package taproot

import (
	"encoding/json"
	"fmt"
	"highgrav/taproot/v1/acacia"
	"highgrav/taproot/v1/authn"
	"net/http"
)

// This assumes that we've already injected the realm, domain, and user into the request.
// DO NOT call this unless you have already done so!
func (srv *AppServer) HandleAcacia(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		realm := r.Context().Value(HTTP_CONTEXT_REALM_KEY).(string)
		dom := r.Context().Value(HTTP_CONTEXT_DOMAIN_KEY).(string)
		usr := r.Context().Value(HTTP_CONTEXT_USER_KEY).(authn.User)
		rr := acacia.NewRightsRequest(realm, dom, usr, r)
		js, err := json.Marshal(rr)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(js))

		next.ServeHTTP(w, r)
	})
}
