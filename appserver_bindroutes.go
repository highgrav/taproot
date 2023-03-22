package taproot

import (
	"fmt"
	"github.com/google/deck"
	"github.com/justinas/alice"
	"net/http"
)

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

	x := 0
	dmw := alice.New()
	for _, rb := range srv.routes {
		deck.Info("Setting route " + rb.Route)
		srv.Router.Handler(rb.Method, rb.Route, dmw.Then(srv.handleLocalMetrics(rb.Handler)))
		x++
	}
	deck.Info(fmt.Sprintf("%d routes added", x))

	// Standard middleware
	defaultMiddleware := []alice.Constructor{
		srv.HandlePanic,
		srv.handleIPFiltering,
		srv.HandleGlobalRateLimit,
		srv.HandleIPRateLimit,
		srv.HandleGlobalMetrics,
		srv.HandleTracing,
		srv.Session.LoadAndSave,
		srv.CreateHandleSession(srv.Config.UseEncryptedSessionTokens),
	}
	// Add any additional custom middleware
	mw := alice.New(append(defaultMiddleware, srv.Middleware...)...) // We always filter
	mw.Append(srv.Middleware...)
	return mw.Then(srv.Router)
}
