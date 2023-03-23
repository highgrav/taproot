package session

import (
	"context"
	"errors"
	"time"
)

var (
	ErrStoreError          = errors.New("error in store")
	ErrKeyNotInSession     = errors.New("key was not found in session")
	ErrTypeCoercionFailure = errors.New("could not coerce stored value to requested type")
)

// We essentially replicate the Store interface from SCS, so we can take advantage of existing implementations
type IStore interface {
	Delete(token string) error
	Find(token string) (b []byte, found bool, err error)
	Commit(token string, b []byte, expiry time.Time) error
}

type IIterableStore interface {
	All() (map[string][]byte, error)
}

type ICtxStore interface {
	IStore
	DeleteCtx(ctx context.Context, token string) (err error)
	FindCtx(ctx context.Context, token string) (b []byte, found bool, err error)
	CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) (err error)
}

type IIterableCtxStore interface {
	AllCtx(ctx context.Context) (map[string][]byte, error)
}
