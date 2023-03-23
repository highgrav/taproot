package taproot

import (
	"context"
	"fmt"
	"github.com/justinas/alice"
	"highgrav/taproot/v1/authn"
	"highgrav/taproot/v1/common"
	"highgrav/taproot/v1/logging"
	"net/http"
	"time"
)

const SESSION_HEADER_KEY string = "X-Session"
const SESSION_EXPIRATION_HEADER_KEY string = "X-Session-Expires-At"
const SESSION_COOKIE_NAME string = "SessionInfo"

/*
HandleSession() checks to see if there is a valid session token in either the cookie or the header, and tries to
rehydrate the session from there.
*/

func (srv *AppServer) CreateHandleSession(encryptTokens bool) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			var user authn.User
			var cookieVal string
			var headerVal string
			var token authn.AuthToken
			cookie, err := r.Cookie(SESSION_COOKIE_NAME)
			if err != nil {
				cookieVal = ""
			} else {
				cookieVal = cookie.Value
			}
			headerVal = r.Header.Get(SESSION_HEADER_KEY)

			// If we don't see anything here, then just inject an anonymous user and move on
			if headerVal == "" && cookieVal == "" {
				ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, authn.Anonymous())
				ctx = context.WithValue(ctx, HTTP_CONTEXT_REALM_KEY, srv.Config.DefaultRealm)
				ctx = context.WithValue(ctx, HTTP_CONTEXT_DOMAIN_KEY, srv.Config.DefaultDomain)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			var tokenVal string = ""
			if headerVal != "" {
				tokenVal = headerVal
			} else if cookieVal != "" {
				// Use cookies if the header isn't set. Header value always take precedence
				tokenVal = cookieVal
			}
			if encryptTokens {
				token, err = srv.SignatureMgr.DecryptToken(tokenVal)
				if err != nil {
					logging.LogToDeck("error", fmt.Sprintf("SESS\tDecrypt Header Token: %s", err.Error()))
					ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, authn.Anonymous())
					ctx = context.WithValue(ctx, HTTP_CONTEXT_REALM_KEY, srv.Config.DefaultRealm)
					ctx = context.WithValue(ctx, HTTP_CONTEXT_DOMAIN_KEY, srv.Config.DefaultDomain)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			} else {
				token, err = srv.SignatureMgr.VerifySignedToken(tokenVal)
				if err != nil {
					logging.LogToDeck("error", fmt.Sprintf("SESS\tVerify Header Token: %s", err.Error()))
					ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, authn.Anonymous())
					ctx = context.WithValue(ctx, HTTP_CONTEXT_REALM_KEY, srv.Config.DefaultRealm)
					ctx = context.WithValue(ctx, HTTP_CONTEXT_DOMAIN_KEY, srv.Config.DefaultDomain)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			if time.Now().After(token.ExpiresAt) {
				// TODO -- return warning to user that their session needs refreshing?s
				srv.Session.Remove(token.Token)
				ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, authn.Anonymous())
				ctx = context.WithValue(ctx, HTTP_CONTEXT_REALM_KEY, srv.Config.DefaultRealm)
				ctx = context.WithValue(ctx, HTTP_CONTEXT_DOMAIN_KEY, srv.Config.DefaultDomain)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			if err != nil {
				logging.LogToDeck("error", "SESS\terror\tError loading SCS session: "+err.Error())
			}
			user, err = srv.GetUserFromSession(token.Token)
			if err != nil {
				logging.LogToDeck("error", fmt.Sprintf("SESS\tError casting session data to user for token %s: %s", token.Token, err.Error()))
				ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, authn.Anonymous())
				ctx = context.WithValue(ctx, HTTP_CONTEXT_REALM_KEY, srv.Config.DefaultRealm)
				ctx = context.WithValue(ctx, HTTP_CONTEXT_DOMAIN_KEY, srv.Config.DefaultDomain)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			ctx = context.WithValue(ctx, HTTP_CONTEXT_USER_KEY, user)
			realmId := user.RealmID
			if realmId == "" {
				user.RealmID = srv.Config.DefaultRealm
				realmId = srv.Config.DefaultRealm
			}
			domainId := user.DomainID
			if domainId == "" {
				user.DomainID = srv.Config.DefaultDomain
				domainId = srv.Config.DefaultDomain
			}
			ctx = context.WithValue(ctx, HTTP_CONTEXT_REALM_KEY, realmId)
			ctx = context.WithValue(ctx, HTTP_CONTEXT_DOMAIN_KEY, domainId)
			bw := &common.BufferedHttpResponseWriter{ResponseWriter: w}
			sr := r.WithContext(ctx)
			next.ServeHTTP(bw, sr)

			// Inject cookie or header data if necessary (we use the buffered response writer so we can inject headers prior to writing the response)

			w.Write(bw.Buf.Bytes())
		})
	}
}
