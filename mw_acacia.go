package taproot

import (
	"context"
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

		rights, err := srv.acacia.Apply(rr)
		if err != nil {
			srv.ErrorResponse(w, r, 500, err.Error())
			return
		}

		// TODO

		// If we have a response, that takes priority

		// If we have a redirect, that takes secondary priority

		// Add rights into the context
		ctx := context.WithValue(r.Context(), HTTP_CONTEXT_ACACIA_RIGHTS_KEY, rights.Rights)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
