package taproot

import "net/http"

// Generates a new WebServer with a mux that redirects to HTTPS.
// TODO -- How does this play with autocert's HTTP server?
func (srv *AppServer) NewHttpRedirectServer(cfg HttpConfig) *WebServer {

	ws := NewWebServer(nil, cfg)

	handleRedirectToHttps := func(w http.ResponseWriter, r *http.Request) {
		// TODO -- strip port from request and add in main webserver port
		newURI := "https://" + r.Host + r.URL.String()
		http.Redirect(w, r, newURI, http.StatusFound)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRedirectToHttps)
	ws.Handler = mux
	return ws
}
