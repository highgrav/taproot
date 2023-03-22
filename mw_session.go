package taproot

import (
	"context"
	"github.com/alexedwards/scs/v2"
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
HandleCookieSession() creates middleware that injects and serializes user data from the request or HTTP context using cookies. encryptCookie() == true means the cookie will be encrypted; otherwise, it will be signed but not encrypted.
*/
func (srv *AppServer) HandleCookieSession(redirectTo string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
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
		next(w, r.WithContext(ctx))
	})
}

/*
HandleHeaderSession() creates middleware that injects and serializes user data from the request or HTTP context using headers. encryptCookie() == true means the header value will be encrypted; otherwise, it will be signed but not encrypted.
*/
func (srv *AppServer) HandleHeaderSession(redirectTo string, next http.Handler) http.Handler {
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
		ctx = context.WithValue(r.Context(), HTTP_CONTEXT_USER_KEY, usr)

		next.ServeHTTP(w, r.WithContext(ctx))

		ctx, err = srv.Session.Load(r.Context(), sesKey)
		if err != nil {
			logging.LogToDeck("error", "SESS\terror\t"+err.Error())
			srv.ErrorResponse(w, r, http.StatusInternalServerError, "internal error")
			return
		}

		// We use a buffered writer here so we can add headers on the way back
		bw := &common.BufferedResponseWriter{ResponseWriter: w}
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
