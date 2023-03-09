package taproot

import (
	"context"
	"fmt"
	"github.com/google/deck"
	"github.com/jpillora/ipfilter"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/time/rate"
	"highgrav/taproot/v1/authn"
	"net"
	"net/http"
	"os"
	"time"
)

type WebServer struct {
	Config       HttpConfig
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

func (srv *WebServer) Close() error {
	return srv.Server.Close()
}

func (srv *WebServer) ListenAndServe() error {
	srv.state.setState(SERVER_STATE_RUNNING)
	return srv.Server.ListenAndServe()
}

func (srv *WebServer) ListenAndServeTLS(certFile, keyFile string) error {
	srv.state.setState(SERVER_STATE_INITIALIZING)

	if srv.Config.TLS.UseSelfSignedCert && !srv.Config.TLS.UseACME {
		deck.Info("Generating self-signed certificate for serving...")
		c, err := generateSelfSignedTlsCert()
		if err != nil {
			return err
		}
		srv.Server.TLSConfig = c
		deck.Info("Serving self-signed TLS on port ", srv.Config.Port)
		return srv.Server.ListenAndServeTLS(certFile, keyFile)
	}

	if srv.Config.TLS.UseACME {

	} else {
		// Ignore ACME, use the provided key files
	}

	srv.state.setState(SERVER_STATE_RUNNING)

	return srv.Server.ListenAndServeTLS(certFile, keyFile)
}

func (srv *WebServer) RegisterOnShutdown(f func()) {
	srv.Server.RegisterOnShutdown(f)
}

func (srv *WebServer) Serve(l net.Listener) error {
	srv.state.setState(SERVER_STATE_RUNNING)
	return srv.Server.Serve(l)
}

func (srv *WebServer) ServeTLS(l net.Listener, certFile, keyFile string) error {
	srv.state.setState(SERVER_STATE_INITIALIZING)

	if srv.Config.TLS.UseSelfSignedCert && !srv.Config.TLS.UseACME {
		deck.Info("Generating self-signed certificate for serving...")
		c, err := generateSelfSignedTlsCert()
		fmt.Printf("Cert count: %d\n", len(srv.Server.TLSConfig.Certificates))

		if err != nil {
			deck.Fatal(err)
			os.Exit(-222)
		}
		srv.Server.TLSConfig = c
		return srv.Server.ServeTLS(l, "", "")
	}

	if srv.Config.TLS.UseACME {

	} else {
		// Ignore ACME, use the provided key files
	}

	srv.state.setState(SERVER_STATE_RUNNING)
	return srv.Server.ServeTLS(l, certFile, keyFile)
}

func (srv *WebServer) SetKeepAlivesEnabled(v bool) {
	srv.Server.SetKeepAlivesEnabled(v)
}

func (srv *WebServer) Shutdown(ctx context.Context) error {
	return srv.Server.Shutdown(ctx)
}