package taproot

import (
	"context"
	"fmt"
	"github.com/google/deck"
	"highgrav/taproot/v1/jsrun"
	"net"
	"os"
)

func (srv *AppServer) AddJSInjector(injectorFunc jsrun.InjectorFunc) {
	srv.jsinjections = append(srv.jsinjections, injectorFunc)
}

func (srv *AppServer) AddMiddleware(middlewareFunc MiddlewareFunc) {
	srv.Middleware = append(srv.Middleware, middlewareFunc)
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
