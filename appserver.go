package taproot

import (
	"database/sql"
	"github.com/highgrav/taproot/v1/acacia"
	"github.com/highgrav/taproot/v1/authn"
	"github.com/highgrav/taproot/v1/cron"
	"github.com/highgrav/taproot/v1/jsrun"
	"github.com/highgrav/taproot/v1/session"
	"github.com/highgrav/taproot/v1/sse"
	"github.com/highgrav/taproot/v1/websock"
	"github.com/highgrav/taproot/v1/workers"
	"github.com/jpillora/ipfilter"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/time/rate"
	"net/http"
	"time"
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
	Session         *session.SessionManager
	Config          ServerConfig
	Router          *httprouter.Router
	Middleware      []alice.Constructor // Used when adding a new route
	DBs             map[string]*sql.DB
	ExitServerCh    chan bool

	SSEHubs      map[string]*sse.SSEHub
	WSHubs       map[string]*websock.WSHub
	WorkHub      *workers.WorkQueue
	CronHub      *cron.CronHub
	SignatureMgr *authn.AuthSignerManager

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
	startedOn         time.Time
	autocert          *autocert.Manager
}
