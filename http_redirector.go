package taproot

import "net/http"

func (srv *Server) startHttpsRedirector() error {
	redirect := func(w http.ResponseWriter, r *http.Request) {
		tgt := "https://"
		http.Redirect(w, r, tgt, http.StatusFound)
	}
	return http.ListenAndServe(":80", http.HandlerFunc(redirect))
}
