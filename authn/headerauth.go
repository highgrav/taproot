package authn

import (
	"context"
	"net/http"
)

type IAuthenticationReader interface {
	WithContext(ctx context.Context, r *http.Request) context.Context
}
