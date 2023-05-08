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

			// TODO -- we'll hardcode some mimetypes for the moment, should move
			// to a configuration option
			if strings.HasSuffix(r.URL.Path, ".js") {
				w.Header().Set("Content-Type", "text/javascript")
			} else if strings.HasSuffix(r.URL.Path, ".json") {
				w.Header().Set("Content-Type", "application/json")
			} else if strings.HasSuffix(r.URL.Path, ".htm") || strings.HasSuffix(r.URL.Path, ".html") {
				w.Header().Set("Content-Type", "text/html")
			} else if strings.HasSuffix(r.URL.Path, ".css") {
				w.Header().Set("Content-Type", "text/css")
			} else if strings.HasSuffix(r.URL.Path, ".xml") {
				w.Header().Set("Content-Type", "application/xml")
			} else if strings.HasSuffix(r.URL.Path, ".xhtml") {
				w.Header().Set("Content-Type", "application/xhtml+xml")
			} else if strings.HasSuffix(r.URL.Path, ".svg") {
				w.Header().Set("Content-Type", "image/svg+xml")
			} else if strings.HasSuffix(r.URL.Path, ".ttf") {
				w.Header().Set("Content-Type", "font/ttf")
			} else if strings.HasSuffix(r.URL.Path, ".otf") {
				w.Header().Set("Content-Type", "font/otf")
			} else if strings.HasSuffix(r.URL.Path, ".jsonld") {
				w.Header().Set("Content-Type", "application/ld+json")
			} else if strings.HasSuffix(r.URL.Path, ".bmp") {
				w.Header().Set("Content-Type", "image/bmp")
			} else if strings.HasSuffix(r.URL.Path, ".csv") {
				w.Header().Set("Content-Type", "text/csv")
			} else if strings.HasSuffix(r.URL.Path, ".gif") {
				w.Header().Set("Content-Type", "image/gif")
			} else if strings.HasSuffix(r.URL.Path, ".jpeg") || strings.HasSuffix(r.URL.Path, ".jpg") {
				w.Header().Set("Content-Type", "image/jpeg")
			} else if strings.HasSuffix(r.URL.Path, ".mp3") {
				w.Header().Set("Content-Type", "audio/mp3")
			} else if strings.HasSuffix(r.URL.Path, ".mp4") {
				w.Header().Set("Content-Type", "video/mp4")
			} else if strings.HasSuffix(r.URL.Path, ".mpeg") {
				w.Header().Set("Content-Type", "video/mpeg")
			} else if strings.HasSuffix(r.URL.Path, ".png") {
				w.Header().Set("Content-Type", "image/png")
			} else if strings.HasSuffix(r.URL.Path, ".txt") {
				w.Header().Set("Content-Type", "text/plain")
			} else if strings.HasSuffix(r.URL.Path, ".webp") {
				w.Header().Set("Content-Type", "image/webp")
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
