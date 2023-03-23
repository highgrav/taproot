package taproot

import (
	"context"
	"fmt"
	"github.com/google/deck"
	"github.com/highgrav/taproot/v1/jsrun"
	"github.com/justinas/alice"
	"net"
	"net/http"
	"os"
)

// Adds a custom function that will be run to inject new objects or functions into the server-side JS runtime.
func (srv *AppServer) AddJSInjector(injectorFunc jsrun.InjectorFunc) {
	srv.jsinjections = append(srv.jsinjections, injectorFunc)
}

// Adds global middleware to all routes.
func (srv *AppServer) AddMiddleware(middlewareFunc alice.Constructor) {
	srv.Middleware = append(srv.Middleware, middlewareFunc)
}

// Adds a route to the app server that will be bound via bindRoutes() when the webserver is started.
func (srv *AppServer) Handler(method string, route string, handler http.Handler) {
	if srv.routes == nil {
		srv.routes = make([]RouteBinding, 0)
	}
	srv.routes = append(srv.routes, RouteBinding{
		Method:  method,
		Route:   route,
		Handler: handler,
	})
}

/*
Close immediately closes all active net.Listeners and any connections in state StateNew, StateActive, or StateIdle. For a graceful shutdown, use Shutdown.

Close does not attempt to close (and does not even know about) any hijacked connections, such as WebSockets.

Close returns any error returned from closing the Server's underlying Listener(s).
*/
func (srv *AppServer) Close() error {
	return srv.Server.Close()
}

/*
Starts the web server in HTTP mode.

ListenAndServe listens on the TCP network address srv.Addr and then calls Serve to handle requests on incoming connections. Accepted connections are configured to enable TCP keep-alives.

If srv.Addr is blank, ":http" is used.

ListenAndServe always returns a non-nil error. After Shutdown or Close, the returned error is ErrServerClosed.
*/
func (srv *AppServer) ListenAndServe() error {
	srv.Server.Server.Handler = srv.bindRoutes()
	srv.state.setState(SERVER_STATE_RUNNING)
	return srv.Server.ListenAndServe()
}

/*
Starts the server in the configured HTTPS mode, preferring ACME, otherwise using an internally-generated self-signed cert,
or key/cert file pairs, depending on configuration. This will also start the HTTP->HTTPS redirect server, if configured
to do so.
*/
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
		srv.state.setState(SERVER_STATE_RUNNING)
		return srv.Server.ListenAndServeTLS(certFile, keyFile)
	}

	if srv.Config.HttpServer.TLS.UseACME {
		srv.startACME()
		srv.state.setState(SERVER_STATE_RUNNING)
		return srv.Server.ListenAndServeTLS(certFile, keyFile)
	} else {
		// Ignore ACME, use the provided key files
		// TODO
	}

	srv.state.setState(SERVER_STATE_RUNNING)
	return srv.Server.ListenAndServeTLS(certFile, keyFile)
}

/*
RegisterOnShutdown registers a function to call on Shutdown. This can be used to gracefully shutdown connections that have undergone ALPN protocol upgrade or that have been hijacked. This function should start protocol-specific graceful shutdown, but should not wait for shutdown to complete.
*/
func (srv *AppServer) RegisterOnShutdown(f func()) {
	srv.Server.RegisterOnShutdown(f)
}

/*
Starts the web server in HTTP mode.

Serve accepts incoming connections on the Listener l, creating a new service goroutine for each. The service goroutines read requests and then call srv.Handler to reply to them.

HTTP/2 support is only enabled if the Listener returns *tls.Conn connections and they were configured with "h2" in the TLS Config.NextProtos.

Serve always returns a non-nil error and closes l. After Shutdown or Close, the returned error is ErrServerClosed.
*/
func (srv *AppServer) Serve(l net.Listener) error {

	srv.Server.Server.Handler = srv.bindRoutes()
	srv.state.setState(SERVER_STATE_RUNNING)
	return srv.Server.Serve(l)
}

/*
Starts the server in the configured HTTPS mode, preferring ACME, otherwise using an internally-generated self-signed cert, or key/cert file pairs.

ServeTLS accepts incoming connections on the Listener l, creating a new service goroutine for each. The service goroutines perform TLS setup and then read requests, calling srv.Handler to reply to them.

Files containing a certificate and matching private key for the server must be provided if neither the Server's TLSConfig.Certificates nor TLSConfig.GetCertificate are populated. If the certificate is signed by a certificate authority, the certFile should be the concatenation of the server's certificate, any intermediates, and the CA's certificate.

ServeTLS always returns a non-nil error. After Shutdown or Close, the returned error is ErrServerClosed.
*/
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
		srv.state.setState(SERVER_STATE_RUNNING)
		return srv.Server.ServeTLS(l, "", "")
	}

	if srv.Config.HttpServer.TLS.UseACME {
		srv.startACME()
		srv.state.setState(SERVER_STATE_RUNNING)
		return srv.Server.ServeTLS(l, "", "")
	} else {
		// Ignore ACME, use the provided key files
		// TODO
	}

	srv.state.setState(SERVER_STATE_RUNNING)
	return srv.Server.ServeTLS(l, certFile, keyFile)
}

// SetKeepAlivesEnabled controls whether HTTP keep-alives are enabled. By default, keep-alives are always enabled. Only very resource-constrained environments or servers in the process of shutting down should disable them.
func (srv *AppServer) SetKeepAlivesEnabled(v bool) {
	srv.Server.SetKeepAlivesEnabled(v)
}

/*
Shutdown gracefully shuts down the web server without interrupting any active connections. Shutdown works by first closing all open listeners, then closing all idle connections, and then waiting indefinitely for connections to return to idle and then shut down. If the provided context expires before the shutdown is complete, Shutdown returns the context's error, otherwise it returns any error returned from closing the Server's underlying Listener(s).

When Shutdown is called, Serve, ListenAndServe, and ListenAndServeTLS immediately return ErrServerClosed. Make sure the program doesn't exit and waits instead for Shutdown to return.

Shutdown does not attempt to close nor wait for hijacked connections such as WebSockets. The caller of Shutdown should separately notify such long-lived connections of shutdown and wait for them to close, if desired. See RegisterOnShutdown for a way to register shutdown notification functions.

Once Shutdown has been called on a server, it may not be reused; future calls to methods such as Serve will return ErrServerClosed.
*/
func (srv *AppServer) Shutdown(ctx context.Context) error {
	return srv.Server.Shutdown(ctx)
}

/*
Adds an endpoint handler wrapped in an Acacia policy tester.
Only endpoints wrapped with WithPolicy() or WithPolicyFunc() will have Acacia policies applied to them.
*/
func (srv *AppServer) WithPolicy(next http.Handler) http.Handler {
	return srv.handleAcacia(next)
}

/*
Adds an endpoint handler, using a function wrapper, wrapped in an Acacia policy tester.
Only endpoints wrapped with WithPolicy() or WithPolicyFunc() will have Acacia policies applied to them.
*/
func (srv *AppServer) WithPolicyFunc(next http.HandlerFunc) http.Handler {
	return srv.handleAcacia(next)
}
