package taproot

import (
	"context"
	"highgrav/taproot/v1/common"
	"net/http"
)

func (srv *Server) HandleLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// time, remote address, method, url, correlation ID, username
		var corrId string = common.CreateRandString(32)

		// Save the correlation ID to context so we can propagate it
		ctx := context.WithValue(r.Context(), CONTEXT_CORRELATION_KEY_NAME, corrId)

		next.ServeHTTP(w, r.WithContext(ctx))

		// Everything below here will be executed on the way back up the chain
		// TODO ...
	})
}
