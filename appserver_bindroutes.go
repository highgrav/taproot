package taproot

import (
	"context"
	"fmt"
	"github.com/highgrav/taproot/logging"
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
		logging.LogToDeck(context.Background(), "info", "TAPROOT", "startup", "Setting route "+rb.Route)
		if rb.Method != "" {
			srv.Router.Handler(rb.Method, rb.Route, dmw.Then(srv.handleLocalMetrics(rb.Handler)))
		} else {
			// TODO
		}
		x++
	}
	logging.LogToDeck(context.Background(), "info", "TAPROOT", "startup", fmt.Sprintf("%d routes added", x))

	// Standard middleware
	defaultMiddleware := []alice.Constructor{
		srv.HandlePanic,
		srv.handleIPFiltering,
		srv.HandleGlobalRateLimit,
		srv.HandleIPRateLimit,
		srv.HandleGlobalMetrics,
		srv.HandleTracing,
		srv.CreateHandleSession(srv.Config.UseEncryptedSessionTokens),
		srv.HandleFeatureFlags,
	}
	// Add any additional custom middleware
	mw := alice.New(append(defaultMiddleware, srv.Middleware...)...) // We always filter
	mw.Append(srv.Middleware...)
	return mw.Then(srv.Router)
}
