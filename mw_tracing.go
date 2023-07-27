package taproot

import (
	"github.com/highgrav/taproot/common"
	"github.com/highgrav/taproot/constants"
	"golang.org/x/net/context"
	"net/http"
)

func (srv *AppServer) HandleTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// time, remote address, method, url, correlation ID, username
		var corrId string = common.CreateRandString(16)

		// Save the correlation ID to context so we can propagate it
		ctx := context.WithValue(r.Context(), constants.HTTP_CONTEXT_CORRELATION_KEY, corrId)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
