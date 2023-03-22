package taproot

import (
	"context"
	"errors"
	"github.com/alexedwards/scs/v2"
	"highgrav/taproot/v1/authn"
	"highgrav/taproot/v1/common"
	"highgrav/taproot/v1/logging"
	"net/http"
)

var ErrSessionKeyExists = errors.New("session key already exists")
var ErrSessionKeyDoesNotExist = errors.New("session key does not exist")
var ErrSessionInvalidType = errors.New("attempt to cast to invalid type")

func GetSessionItem[T any](ses *scs.SessionManager, ctx context.Context, key string) (T, error) {
	var t T
	if !ses.Exists(ctx, key) {
		return t, ErrSessionKeyDoesNotExist
	}
	tmpVal := ses.Get(ctx, key)
	t, ok := tmpVal.(T)
	if !ok {
		return t, ErrSessionInvalidType
	}
	return t, nil
}

func (svr *AppServer) AddUserToSession(ctx context.Context, user authn.User) (string, error) {
	key := common.CreateRandString(16)
	for svr.Session.Exists(ctx, key) {
		key = common.CreateRandString(16)
	}
	return key, svr.AddSession(ctx, key, user)
}

func (svr *AppServer) AddOrReplaceUserToSession(ctx context.Context, key string, user authn.User) (string, error) {
	svr.AddOrReplaceSession(ctx, key, user)
	return key, nil
}

func (svr *AppServer) RemoveSession(ctx context.Context, key string) {
	svr.Session.Remove(ctx, key)
}

func (svr *AppServer) AddSession(ctx context.Context, key string, t any) error {
	if svr.Session.Exists(ctx, key) {
		return ErrSessionKeyExists
	}
	svr.Session.Put(ctx, key, t)
	return nil
}

func (svr *AppServer) AddOrReplaceSession(ctx context.Context, key string, t any) {
	svr.Session.Put(ctx, key, t)
}

func (svr *AppServer) ReplaceSession(ctx context.Context, key string, t any) error {
	if !svr.Session.Exists(ctx, key) {
		return ErrSessionKeyDoesNotExist
	}
	svr.Session.Put(ctx, key, t)
	return nil
}

func (srv *AppServer) handleSessionError(w http.ResponseWriter, r *http.Request, err error) {
	logging.LogToDeck("error", "SESS\terror\t"+err.Error())
	srv.ErrorResponse(w, r, 500, "session error encountered")
}
