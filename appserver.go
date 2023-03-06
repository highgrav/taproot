package taproot

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/google/deck"
	"github.com/google/deck/backends/logger"
	"github.com/jpillora/ipfilter"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/time/rate"
	"highgrav/taproot/v1/acacia"
	"highgrav/taproot/v1/authn"
	"highgrav/taproot/v1/jsrun"
	"net"
	"net/http"
	"os"
	"time"
)

type MiddlewareFunc func(http.Handler) http.Handler

type AppServer struct {
	Session      *scs.SessionManager
	Config       ServerConfig
	Server       *http.Server // Main HTTP Server
	Router       *httprouter.Router
	Middleware   []MiddlewareFunc // Used when adding a new route
	DBs          map[string]*sql.DB
	ExitServerCh chan bool

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
}

func New(userStore authn.IUserStore, cfg ServerConfig) *AppServer {
	// set up logging (we use stdout until the server is up and running)
	deck.Add(logger.Init(os.Stdout, 0))

	s := &AppServer{}
	s.Config = cfg
	s.users = userStore
	s.DBs = make(map[string]*sql.DB)
	s.Middleware = make([]MiddlewareFunc, 0)
	s.jsinjections = make([]jsrun.InjectorFunc, 0)

	// Set up IP filter
	// TODO
	s.httpIpFilter = ipfilter.New(ipfilter.Options{})

	// Set up our feature flags
	// TODO

	// Set up our security policy authorizer
	sa, err := acacia.New(cfg.SecurityPolicyDir)
	if err != nil {
		deck.Fatal(err.Error())
		os.Exit(-1)
	}
	s.authz = sa

	// set up our JS manager
	js, err := jsrun.New(cfg.ScriptFilePath)
	if err != nil {
		deck.Fatal(err.Error())
		os.Exit(-1)
	}
	s.js = js

	if s.Config.UseJSMLFiles {
		err = s.compileJSMLFiles(s.Config.JSMLFilePath, s.Config.JSMLCompiledFilePath)
		if err != nil {
			deck.Fatal(err.Error())
			os.Exit(-1)
		}
	}

	s.Router = httprouter.New()
	s.Router.SaveMatchedRoutePath = true // necessary to get the matched path back for Acacia authz
	s.Server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HttpServer.ServerName, cfg.HttpServer.Port),
		Handler:      s.Router,
		IdleTimeout:  time.Duration(cfg.HttpServer.Timeouts.Idle) * time.Second,
		ReadTimeout:  time.Duration(cfg.HttpServer.Timeouts.Read) * time.Second,
		WriteTimeout: time.Duration(cfg.HttpServer.Timeouts.Write) * time.Second,
	}

	return s
}

func (srv *AppServer) AddJSInjector(injectorFunc jsrun.InjectorFunc) {
	srv.jsinjections = append(srv.jsinjections, injectorFunc)
}

func (srv *AppServer) AddMiddleware(middlewareFunc MiddlewareFunc) {
	srv.Middleware = append(srv.Middleware, middlewareFunc)
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

/* http.Server overloads */
func (srv *AppServer) Close() error {
	return srv.Server.Close()
}

func (srv *AppServer) ListenAndServe() error {
	srv.Server.Handler = srv.bindRoutes()
	srv.state.setState(SERVER_STATE_RUNNING)
	return srv.Server.ListenAndServe()
}

func (srv *AppServer) ListenAndServeTLS(certFile, keyFile string) error {
	srv.Server.Handler = srv.bindRoutes()
	srv.state.setState(SERVER_STATE_INITIALIZING)

	if srv.Config.HttpServer.TLS.UseSelfSignedCert && !srv.Config.HttpServer.TLS.UseACME {
		deck.Info("Generating self-signed certificate for serving...")
		c, err := srv.generateSelfSignedTlsCert()
		if err != nil {
			return err
		}
		srv.Server.TLSConfig = c
		deck.Info("Serving self-signed TLS on port ", srv.Config.HttpServer.Port)
		return srv.Server.ListenAndServeTLS(certFile, keyFile)
	}

	if srv.Config.HttpServer.TLS.UseACME {

	} else {
		// Ignore ACME, use the provided key files
	}

	srv.state.setState(SERVER_STATE_RUNNING)

	return srv.Server.ListenAndServeTLS(certFile, keyFile)
}

func (srv *AppServer) RegisterOnShutdown(f func()) {
	srv.Server.RegisterOnShutdown(f)
}

func (srv *AppServer) Serve(l net.Listener) error {
	srv.Server.Handler = srv.bindRoutes()
	srv.state.setState(SERVER_STATE_RUNNING)
	return srv.Server.Serve(l)
}

func (srv *AppServer) ServeTLS(l net.Listener, certFile, keyFile string) error {
	srv.Server.Handler = srv.bindRoutes()
	srv.state.setState(SERVER_STATE_INITIALIZING)

	if srv.Config.HttpServer.TLS.UseSelfSignedCert && !srv.Config.HttpServer.TLS.UseACME {
		deck.Info("Generating self-signed certificate for serving...")
		c, err := srv.generateSelfSignedTlsCert()
		fmt.Printf("Cert count: %d\n", len(srv.Server.TLSConfig.Certificates))

		if err != nil {
			deck.Fatal(err)
			os.Exit(-222)
		}
		srv.Server.TLSConfig = c
		return srv.Server.ServeTLS(l, "", "")
	}

	if srv.Config.HttpServer.TLS.UseACME {

	} else {
		// Ignore ACME, use the provided key files
	}

	srv.state.setState(SERVER_STATE_RUNNING)
	return srv.Server.ServeTLS(l, certFile, keyFile)
}

func (srv *AppServer) SetKeepAlivesEnabled(v bool) {
	srv.Server.SetKeepAlivesEnabled(v)
}

func (srv *AppServer) Shutdown(ctx context.Context) error {
	return srv.Server.Shutdown(ctx)
}
