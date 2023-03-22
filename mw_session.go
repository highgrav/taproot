package taproot

import (
	"context"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/justinas/alice"
	"highgrav/taproot/v1/authn"
	"highgrav/taproot/v1/common"
	"highgrav/taproot/v1/logging"
	"net/http"
	"time"
)

const SessionHeaderKey string = "X-Session"
const SessionExpirationHeaderKey string = "X-Session-Expires-At"
const CookieSessionKey string = "SessionInfo"

/*
HandleSession() checks to see if there is a valid session token in either the cookie or the header, and tries to
rehydrate the session from there.
*/

func (srv *AppServer) CreateHandleSession(encryptTokens bool) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var ctx context.Context = r.Context()
			var user authn.User
			var cookieVal string
			var headerVal string
			var token authn.AuthToken
			cookie, err := r.Cookie(CookieSessionKey)
			if err != nil {
				cookieVal = ""
			} else {
				cookieVal = cookie.Value
			}
			headerVal = r.Header.Get(SessionHeaderKey)

			// If we don't see anything here, then just inject an anonymous user and move on
			if headerVal == "" && cookieVal == "" {
				ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, authn.Anonymous())
				ctx = context.WithValue(r.Context(), HTTP_CONTEXT_REALM_KEY, srv.Config.DefaultRealm)
				ctx = context.WithValue(r.Context(), HTTP_CONTEXT_DOMAIN_KEY, srv.Config.DefaultDomain)
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
					ctx = context.WithValue(r.Context(), HTTP_CONTEXT_REALM_KEY, srv.Config.DefaultRealm)
					ctx = context.WithValue(r.Context(), HTTP_CONTEXT_DOMAIN_KEY, srv.Config.DefaultDomain)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			} else {
				token, err = srv.SignatureMgr.VerifySignedToken(tokenVal)
				if err != nil {
					logging.LogToDeck("error", fmt.Sprintf("SESS\tVerify Header Token: %s", err.Error()))
					ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, authn.Anonymous())
					ctx = context.WithValue(r.Context(), HTTP_CONTEXT_REALM_KEY, srv.Config.DefaultRealm)
					ctx = context.WithValue(r.Context(), HTTP_CONTEXT_DOMAIN_KEY, srv.Config.DefaultDomain)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			if time.Now().After(token.ExpiresAt) {
				// TODO -- return warning to user that their session needs refreshing?s
				srv.Session.Remove(ctx, token.Token)
				ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, authn.Anonymous())
				ctx = context.WithValue(r.Context(), HTTP_CONTEXT_REALM_KEY, srv.Config.DefaultRealm)
				ctx = context.WithValue(r.Context(), HTTP_CONTEXT_DOMAIN_KEY, srv.Config.DefaultDomain)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			user, ok := srv.Session.Get(ctx, token.Token).(authn.User)
			if !ok {
				logging.LogToDeck("error", fmt.Sprintf("SESS\tError casting session data to user: %s", err.Error()))
				ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, authn.Anonymous())
				ctx = context.WithValue(r.Context(), HTTP_CONTEXT_REALM_KEY, srv.Config.DefaultRealm)
				ctx = context.WithValue(r.Context(), HTTP_CONTEXT_DOMAIN_KEY, srv.Config.DefaultDomain)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, user)
			realmId := user.RealmID
			if realmId == "" {
				realmId = srv.Config.DefaultRealm
			}
			domainId := user.DomainID
			if domainId == "" {
				realmId = srv.Config.DefaultDomain
			}
			ctx = context.WithValue(r.Context(), HTTP_CONTEXT_REALM_KEY, realmId)
			ctx = context.WithValue(r.Context(), HTTP_CONTEXT_DOMAIN_KEY, domainId)
			bw := &common.BufferedHttpResponseWriter{ResponseWriter: w}
			sr := r.WithContext(ctx)
			next.ServeHTTP(bw, sr)

			// Inject cookie or header data if necessary (we use the buffered response writer so we can inject headers prior to writing the response)

			w.Write(bw.Buf.Bytes())
		})
	}
}

/*
HandleCookieSession() creates middleware that injects and serializes user data from the request or HTTP context using cookies. encryptCookie() == true means the cookie will be encrypted; otherwise, it will be signed but not encrypted.
*/
func (srv *AppServer) handleCookieSession(redirectTo string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieSessionKey)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		hVal := cookie.Value
		var ctx context.Context

		if hVal == "" && redirectTo != "" {
			http.Redirect(w, r, redirectTo, 301)
			return
		} else if hVal == "" {
			ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, authn.Anonymous())
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ua, err := srv.SignatureMgr.DecryptToken(hVal)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Retrieve struct from session
		sessionStruct, ok := srv.Session.Get(r.Context(), ua.Token).(authn.User)
		if !ok {
			srv.ErrorResponse(w, r, http.StatusUnauthorized, "invalid authorization")
			return
		}

		// Place struct in request context
		ctx = context.WithValue(r.Context(), "user", sessionStruct)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

/*
HandleHeaderSession() creates middleware that injects and serializes user data from the request or HTTP context using headers. encryptCookie() == true means the header value will be encrypted; otherwise, it will be signed but not encrypted.
*/
func (srv *AppServer) handleHeaderSession(redirectTo string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx context.Context
		// get user/session from header value
		hVal := r.Header.Get(SessionHeaderKey)
		var sesKey string // either a session or user key
		if hVal == "" && redirectTo != "" {
			http.Redirect(w, r, redirectTo, 301)
			return
		} else if hVal == "" {
			ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, authn.Anonymous())
			ctx = context.WithValue(r.Context(), HTTP_CONTEXT_REALM_KEY, srv.Config.DefaultRealm)
			ctx = context.WithValue(r.Context(), HTTP_CONTEXT_DOMAIN_KEY, srv.Config.DefaultDomain)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ua, err := srv.SignatureMgr.DecryptToken(hVal)
		if err != nil {
			srv.ErrorResponse(w, r, http.StatusUnauthorized, "invalid authorization")
			return
		}
		usr, ok := srv.Session.Get(r.Context(), ua.Token).(authn.User)
		if !ok {
			srv.ErrorResponse(w, r, http.StatusInternalServerError, "error retrieving session")
		}
		realmId := usr.RealmID
		if realmId == "" {
			realmId = srv.Config.DefaultRealm
		}
		domainId := usr.DomainID
		if domainId == "" {
			realmId = srv.Config.DefaultDomain
		}
		ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, usr)
		ctx = context.WithValue(r.Context(), HTTP_CONTEXT_REALM_KEY, realmId)
		ctx = context.WithValue(r.Context(), HTTP_CONTEXT_DOMAIN_KEY, domainId)

		fmt.Println("APPLIED CONTEXT")

		next.ServeHTTP(w, r.WithContext(ctx))

		ctx, err = srv.Session.Load(r.Context(), sesKey)
		if err != nil {
			logging.LogToDeck("error", "SESS\terror\t"+err.Error())
			srv.ErrorResponse(w, r, http.StatusInternalServerError, "internal error")
			return
		}

		// We use a buffered writer here so we can add headers on the way back
		bw := &common.BufferedHttpResponseWriter{ResponseWriter: w}
		sr := r.WithContext(ctx)
		next.ServeHTTP(bw, sr)

		// set next session token
		if srv.Session.Status(ctx) == scs.Modified {
			if ua.ExpiresAt.Before(time.Now()) {
				hVal = srv.SignatureMgr.NewEncryptedToken(ua.Token)
				bw.Header().Set(SessionHeaderKey, hVal)
				bw.Header().Set(SessionExpirationHeaderKey, srv.SignatureMgr.CurrentSignatureExpiration.Format(http.TimeFormat))
			} else {
				bw.Header().Set(SessionHeaderKey, hVal)
				bw.Header().Set(SessionExpirationHeaderKey, ua.ExpiresAt.Format(http.TimeFormat))
			}
		}
		if bw.Code != 0 {
			w.WriteHeader(bw.Code)
		}
		w.Write(bw.Buf.Bytes())
	})
}
