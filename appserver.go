package taproot

import (
	"database/sql"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/google/deck"
	"github.com/jpillora/ipfilter"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"golang.org/x/time/rate"
	"highgrav/taproot/v1/acacia"
	"highgrav/taproot/v1/authn"
	"highgrav/taproot/v1/jsrun"
	"highgrav/taproot/v1/sse"
	"highgrav/taproot/v1/websock"
	"net/http"
)

// RouteBinding is used when adding a new route endpoint to the app server. It should not be addressed directly.
type RouteBinding struct {
	Method  string
	Route   string
	Handler http.Handler
}

// AppServer is the core data structure for the embedded application server.
type AppServer struct {
	SiteDisplayName string
	Session         *scs.SessionManager
	Config          ServerConfig
	Router          *httprouter.Router
	Middleware      []alice.Constructor // Used when adding a new route
	DBs             map[string]*sql.DB
	ExitServerCh    chan bool

	SSEHubs map[string]*sse.SSEHub
	WSHubs  map[string]*websock.WSHub

	Server *WebServer
	// These are embedded mini-servers for various admin tasks
	RedirectServer *WebServer // Port 80 Server to redirect to https, if not using TLS
	MetricsServer  *WebServer // Dumps performance metrics
	AdminServer    *WebServer // Allows administration
	Acacia         *acacia.PolicyManager

	js                *jsrun.JSManager
	jsinjections      []jsrun.InjectorFunc
	state             serverStateManager
	users             authn.IUserStore
	globalRateLimiter *rate.Limiter
	ipRateLimiter     map[string]*rate.Limiter
	httpIpFilter      *ipfilter.IPFilter
	fflags            retriever.Retriever
	routes            []RouteBinding
	stats             map[string]stats
	globalStats       stats
}

/*
This function takes  user-added routes and wraps them in additional middleware.
Anything added to the server using server.AddMiddleware() will be wrapped as a
global middleware shared across all routes. bindRoutes() also automatically wraps
each endpoint in handleLocalMetrics() (necessary because we depend on being able to
get the matched route prototype -- we want stats to be collected for /some/:id, not
/some/1234325245.
*/
func (srv *AppServer) bindRoutes() http.Handler {
	srv.Router.SaveMatchedRoutePath = true
	if len(srv.Middleware) == 0 {
		deck.Info("No middleware defined, setting routes")
		x := 0
		for _, rb := range srv.routes {
			srv.Router.Handler(rb.Method, rb.Route, srv.handleLocalMetrics(rb.Handler))
			x++
		}
		deck.Info(fmt.Sprintf("%d routes added", x))
		return srv.Router
	}

	deck.Info("Setting routes")
	x := 0

	dmw := alice.New()
	for _, rb := range srv.routes {
		deck.Info("Setting route " + rb.Route)
		srv.Router.Handler(rb.Method, rb.Route, dmw.Then(srv.handleLocalMetrics(rb.Handler)))
		x++
	}
	deck.Info(fmt.Sprintf("%d routes added", x))
	mw := alice.New(srv.Middleware...)
	return mw.Then(srv.Router)
}
