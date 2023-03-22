package taproot

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"highgrav/taproot/v1/acacia"
	"highgrav/taproot/v1/authn"
	"highgrav/taproot/v1/logging"
	"net/http"
)

/*
This function is automatically wrapped around endpoints created using the WithPolicy() and WithPolicyFunc() functions.
It attempts to match one or more Acacia policies and determine what to do (return with an HTTP code and message; redirect
to a different URL; or return a list of permissions). If no policies match, an empty set of permissions will be returned.
This allows business logic to test if user permissions exist without having to write custom code to manage permissions.

Important note: this function assumes that we've already injected the realm, domain, and user into the request, with the
*http.Request's context containing them at HTTP_CONTEXT_REALM_KEY, HTTP_CONTEXT_DOMAIN_KEY, and HTTP_CONTEXT_USER_KEY.
DO NOT call this unless you have already done so!
*/
func (srv *AppServer) handleAcacia(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var realm string
		var dom string
		var usr authn.User = authn.User{}
		if r.Context().Value(HTTP_CONTEXT_REALM_KEY) != nil {
			realm = r.Context().Value(HTTP_CONTEXT_REALM_KEY).(string)
		}
		if r.Context().Value(HTTP_CONTEXT_DOMAIN_KEY) != nil {
			dom = r.Context().Value(HTTP_CONTEXT_DOMAIN_KEY).(string)
		}
		if r.Context().Value(HTTP_CONTEXT_USER_KEY) != nil {
			usr = r.Context().Value(HTTP_CONTEXT_USER_KEY).(authn.User)
		}

		if realm == "" || dom == "" {
			logging.LogToDeck("error", "ACAC/terror/tMissing domain "+dom+" or realm "+realm)
			srv.ErrorResponse(w, r, 500, "failed to apply security policy")
			return
		}
		//fmt.Printf("%+v\n", usr)
		rr := acacia.NewRightsRequest(realm, dom, usr, r)

		params := httprouter.ParamsFromContext(r.Context())

		rights, err := srv.Acacia.Apply(params.MatchedRoutePath(), rr)
		if err != nil {
			srv.ErrorResponse(w, r, 500, err.Error())
			return
		}

		// If we have a response, that takes priority
		if rights.Type == acacia.RESP_TYPE_RESPONSE {
			logging.LogToDeck("info", "ACAC\tinfo\tReceived a short-circuit response from Acacia")
			srv.ErrorResponse(w, r, rights.Response.ReturnCode, rights.Response.ReturnMsg)
			return
		}
		// If we have a redirect, that takes secondary priority
		if rights.Type == acacia.RESP_TYPE_REDIRECT {
			logging.LogToDeck("info", "ACAC\tinfo\tReceived a redirect from Acacia")
			http.Redirect(w, r, rights.Redirect, http.StatusSeeOther)
			return
		}

		// Add rights into the context
		ctx := context.WithValue(r.Context(), HTTP_CONTEXT_ACACIA_RIGHTS_KEY, rights.Rights)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
