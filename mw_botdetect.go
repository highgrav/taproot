package taproot

import (
	"github.com/highgrav/taproot/v1/logging"
	"github.com/x-way/crawlerdetect"
	"net/http"
)

const (
	BOT_DENY_BOT             uint = 2
	BOT_DENY_NON_CRAWLER     uint = 4
	BOT_REDIRECT_NON_CRAWLER uint = 8
	BOT_REDIRECT_CRAWLER     uint = 16
)

func (srv *AppServer) HandleBotDetect(botFlag uint, redirectTo string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if crawlerdetect.IsCrawler(r.UserAgent()) {
			switch botFlag {
			case BOT_DENY_BOT:
				logging.LogToDeck(r.Context(), "info", "BOT", "info", "denied bot of type "+r.UserAgent())
				srv.ErrorResponse(w, r, http.StatusForbidden, "denied")
				return
			case BOT_REDIRECT_CRAWLER:
				http.Redirect(w, r, redirectTo, http.StatusTemporaryRedirect)
				return
			}
		} else {
			switch botFlag {
			case BOT_DENY_NON_CRAWLER:
				srv.ErrorResponse(w, r, http.StatusNotFound, "not found")
				return
			case BOT_REDIRECT_NON_CRAWLER:
				http.Redirect(w, r, redirectTo, http.StatusPermanentRedirect)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
