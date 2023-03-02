package taproot

import (
	"golang.org/x/net/context"
	"highgrav/taproot/v1/common"
	"net/http"
)

func (srv *Server) HandleAddCorsEverywhereHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

// This middleware injects multiple security headers.
func (srv *Server) HandleAddSecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cspNonce := common.CreateRandString(10)
		cspDetails := "object-src 'none'; script-src 'nonce-" + cspNonce + "' 'unsafe-inline' 'unsafe-eval' 'strict-dynamic' https: http:; base-uri 'none';"
		w.Header().Set("X-XSS-Protection", "1 mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("Content-Security-Policy", cspDetails)
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includesubdomains;")
		ctx := context.WithValue(r.Context(), CONTEXT_CSP_NONCE_KEY_NAME, cspNonce)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (srv *Server) HandleAddJsonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// TODO -- does it make sense for us to nail CORS to the HtptpServer config only?
func (srv *Server) HandleAddCorsTrustedHeaders(next http.Handler) http.Handler {
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
