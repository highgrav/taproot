package taproot

import (
	"fmt"
	"github.com/jpillora/ipfilter"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/time/rate"
	"highgrav/taproot/v1/authn"
	"net/http"
	"time"
)

type WebServer struct {
	Host         string
	Port         int
	Server       *http.Server
	Router       *httprouter.Router
	Middleware   []MiddlewareFunc // Used when adding a new route
	ExitServerCh chan bool

	ipFilter          *ipfilter.IPFilter
	globalRateLimiter *rate.Limiter
	ipRateLimiter     map[string]*rate.Limiter
	state             serverStateManager
	users             authn.IUserStore
}

func NewWebServer(userStore authn.IUserStore, cfg HttpConfig) *WebServer {
	s := &WebServer{}
	s.Middleware = make([]MiddlewareFunc, 0)
	s.users = userStore
	s.ipFilter = ipfilter.New(ipfilter.Options{})

	s.Router = httprouter.New()
	s.Router.SaveMatchedRoutePath = true // necessary to get the matched path back for Acacia authz
	s.Server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.ServerName, cfg.Port),
		Handler:      s.Router,
		IdleTimeout:  time.Duration(cfg.Timeouts.Idle) * time.Second,
		ReadTimeout:  time.Duration(cfg.Timeouts.Read) * time.Second,
		WriteTimeout: time.Duration(cfg.Timeouts.Write) * time.Second,
	}

	return s
}
