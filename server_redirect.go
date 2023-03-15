package taproot

import "net/http"

// Generates a new WebServer with a mux that redirects to HTTPS
func NewHttpRedirectServer(cfg HttpConfig) *WebServer {
	ws := NewWebServer(nil, cfg)

	handleRedirectToHttps := func(w http.ResponseWriter, r *http.Request) {
		newURI := "https://" + r.Host + r.URL.String()
		http.Redirect(w, r, newURI, http.StatusFound)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRedirectToHttps)
	ws.Handler = mux
	return ws
}

func (srv *AppServer) startHttpsRedirector() error {
	redirect := func(w http.ResponseWriter, r *http.Request) {
		tgt := "https://"
		http.Redirect(w, r, tgt, http.StatusFound)
	}
	return http.ListenAndServe(":80", http.HandlerFunc(redirect))
}
