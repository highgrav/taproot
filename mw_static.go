package taproot

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

func (srv *AppServer) HandleStaticFiles(next http.Handler) http.Handler {
	s, err := os.Stat(srv.Config.StaticFilePath)
	if err != nil {
		panic(err)
	}
	if !s.IsDir() {
		panic("Static file directory " + srv.Config.StaticFilePath + " is not a directory")
	}
	staticFs := http.Dir(srv.Config.StaticFilePath)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, srv.Config.StaticUrlPath) {
			// don't return any directory listings
			if strings.HasSuffix(r.URL.Path, "/") {
				http.NotFound(w, r)
				return
			}
			spath := strings.TrimPrefix(r.URL.Path, srv.Config.StaticUrlPath)
			surl, err := url.Parse(spath)
			if err != nil {
				panic(err)
			}
			// TODO -- cache access here
			r2 := r.Clone(r.Context())
			r2.URL = surl
			http.FileServer(staticFs).ServeHTTP(w, r2)
			return
		}
		next.ServeHTTP(w, r)
	})
}
