package taproot

import (
	"database/sql"
	"github.com/alexedwards/scs/v2"
	"github.com/jpillora/ipfilter"
	"github.com/julienschmidt/httprouter"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"golang.org/x/time/rate"
	"highgrav/taproot/v1/acacia"
	"highgrav/taproot/v1/authn"
	"highgrav/taproot/v1/jsrun"
	"highgrav/taproot/v1/sse"
	"net/http"
)

type MiddlewareFunc func(http.Handler) http.Handler

type AppServer struct {
	SiteDisplayName string
	Session         *scs.SessionManager
	Config          ServerConfig
	Router          *httprouter.Router
	Middleware      []MiddlewareFunc // Used when adding a new route
	DBs             map[string]*sql.DB
	ExitServerCh    chan bool

	SseBrokers map[string]*sse.SSEBroker

	Server *WebServer
	// These are embedded mini-servers for various admin tasks
	RedirectServer *WebServer // Port 80 Server to redirect to https, if not using TLS
	MetricsServer  *WebServer // Dumps performance metrics
	AdminServer    *WebServer // Allows administration

	js                *jsrun.JSManager
	jsinjections      []jsrun.InjectorFunc
	state             serverStateManager
	users             authn.IUserStore
	authz             *acacia.PolicyManager
	globalRateLimiter *rate.Limiter
	ipRateLimiter     map[string]*rate.Limiter
	httpIpFilter      *ipfilter.IPFilter
	fflags            retriever.Retriever
}

// This takes the user-added routes and wraps them in additional middleware.
// Note that these aren't bound until the server is started.
func (srv *AppServer) bindRoutes() http.Handler {
	if len(srv.Middleware) == 0 {
		return srv.Router
	}
	var h http.Handler = srv.Router
	for x := len(srv.Middleware) - 1; x >= 0; x-- {
		h = srv.Middleware[x](h)
	}
	return h
}
