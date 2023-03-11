package taproot

import (
	"context"
	"fmt"
	"github.com/google/deck"
	"highgrav/taproot/v1/jsrun"
	"highgrav/taproot/v1/sse"
	"net"
	"os"
)

func (srv *AppServer) AddSSEHub(name string) {
	if srv.SSEHubs == nil {
		srv.SSEHubs = make(map[string]*sse.SSEHub)
	}
	if _, ok := srv.SSEHubs[name]; ok {
		return
	}
	b := sse.New(name)
	srv.SSEHubs[name] = b
}

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
	srv.Server.Server.Handler = srv.bindRoutes()
	srv.state.setState(SERVER_STATE_RUNNING)
	return srv.Server.ListenAndServe()
}

func (srv *AppServer) ListenAndServeTLS(certFile, keyFile string) error {
	srv.Server.Server.Handler = srv.bindRoutes()
	srv.state.setState(SERVER_STATE_INITIALIZING)

	if srv.Config.HttpServer.TLS.UseSelfSignedCert && !srv.Config.HttpServer.TLS.UseACME {
		deck.Info("Generating self-signed certificate for serving...")
		c, err := generateSelfSignedTlsCert()
		if err != nil {
			return err
		}
		srv.Server.Server.TLSConfig = c
		deck.Info("Serving self-signed TLS on port ", srv.Config.HttpServer.Port)
		if srv.Config.UseHttpsRedirectServer {
			// TODO
		}
		return srv.Server.ListenAndServeTLS(certFile, keyFile)
	}

	if srv.Config.HttpServer.TLS.UseACME {
		if srv.Config.UseHttpsRedirectServer {
			// TODO
		}
	} else {
		// Ignore ACME, use the provided key files
		if srv.Config.UseHttpsRedirectServer {
			// TODO
		}
	}

	srv.state.setState(SERVER_STATE_RUNNING)

	return srv.Server.ListenAndServeTLS(certFile, keyFile)
}

func (srv *AppServer) RegisterOnShutdown(f func()) {
	srv.Server.RegisterOnShutdown(f)
}

func (srv *AppServer) Serve(l net.Listener) error {
	srv.Server.Server.Handler = srv.bindRoutes()
	srv.state.setState(SERVER_STATE_RUNNING)
	return srv.Server.Serve(l)
}

func (srv *AppServer) ServeTLS(l net.Listener, certFile, keyFile string) error {
	srv.Server.Server.Handler = srv.bindRoutes()
	srv.state.setState(SERVER_STATE_INITIALIZING)

	if srv.Config.HttpServer.TLS.UseSelfSignedCert && !srv.Config.HttpServer.TLS.UseACME {
		deck.Info("Generating self-signed certificate for serving...")
		c, err := generateSelfSignedTlsCert()
		fmt.Printf("Cert count: %d\n", len(srv.Server.Server.TLSConfig.Certificates))

		if err != nil {
			deck.Fatal(err)
			os.Exit(-222)
		}
		srv.Server.Server.TLSConfig = c
		if srv.Config.UseHttpsRedirectServer {
			// TODO
		}
		return srv.Server.ServeTLS(l, "", "")
	}

	if srv.Config.HttpServer.TLS.UseACME {
		if srv.Config.UseHttpsRedirectServer {
			// TODO
		}
	} else {
		// Ignore ACME, use the provided key files
		if srv.Config.UseHttpsRedirectServer {
			// TODO
		}
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
