package taproot

import (
	"github.com/highgrav/taproot/common"
	"github.com/highgrav/taproot/constants"
	"golang.org/x/net/context"
	"net/http"
)

func (srv *AppServer) HandleAddCorsEverywhereHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var src string = r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Origin", src)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "X-Session,Accept,Content-Type,Dnt,Referer,Sec-Ch-Ua,Sec-Ch-Ua-Mobile,Sec-Ch-Ua-Platform,User-Agent,Accept-Charset,Accept-Datetime,Accept-Encoding,Accept-Language,Authorization,Cache-Control,Cookie,Date,Expect,Forwarded,X-Forwarded-For,X-Forwarded-Host,X-Forwarded-Proto,Pragma")
		w.Header().Set("Access-Control-Expose-Headers", "X-Session")
		next.ServeHTTP(w, r)
	})
}

// This middleware injects multiple security headers.
func (srv *AppServer) HandleAddSecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cspNonce := common.CreateRandString(10)
		cspDetails := "object-src 'none'; script-src 'nonce-" + cspNonce + "' 'unsafe-inline' 'unsafe-eval' 'strict-dynamic' https: http:; base-uri 'none';"
		w.Header().Set("X-XSS-Protection", "1 mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("Content-Security-Policy", cspDetails)
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includesubdomains;")
		ctx := context.WithValue(r.Context(), constants.HTTP_CONTEXT_CSP_NONCE_KEY, cspNonce)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (srv *AppServer) HandleAddJsonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// TODO -- does it make sense for us to nail CORS to the HtptpServer config only?
func (srv *AppServer) HandleAddCorsTrustedHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(srv.Config.HttpServer.CorsDomains) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")
		if origin != "" && origin != "null" {
			for _, v := range srv.Config.HttpServer.CorsDomains {
				if origin == v && v != "null" {
					w.Header().Set("Access-Control-Allow-Origin", v)

					// if it looks like a CORS preflight, then send the expected response and short-circuit
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Request-Headers", "Authorization, Content-Type")
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
