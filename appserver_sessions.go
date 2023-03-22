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
var ErrSessionManagerNotInitialized = errors.New("session manager not initialized")

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

/*
AuthenticateUser() returns an authenticated user, if applicable, returning the user and error.
*/
func (svr *AppServer) AuthenticateUser(ctx context.Context, authReq authn.UserAuth) (authn.User, error) {
	user, err := svr.users.GetUserByAuth(authReq)
	if err != nil {
		logging.LogToDeck("error", "AUTH\terror\tError getting user: "+err.Error())
		return authn.Anonymous(), err
	}
	return user, nil
}

/*
RegisterUser() authenticates a user and creates a new session for them, returning the user, session key, and error.
*/
func (svr *AppServer) RegisterUser(ctx context.Context, authReq authn.UserAuth) (authn.User, string, error) {
	if svr.Session == nil {
		return authn.Anonymous(), "", ErrSessionManagerNotInitialized
	}
	user, err := svr.users.GetUserByAuth(authReq)
	if err != nil {
		logging.LogToDeck("error", "AUTH\terror\tError getting user: "+err.Error())
		return authn.Anonymous(), "", err
	}
	key, err := svr.AddUserToSession(ctx, user)
	if err != nil {
		logging.LogToDeck("error", "AUTH\terror\tError adding user to session: "+err.Error())
		return authn.Anonymous(), "", err
	}
	return user, key, nil
}

func (svr *AppServer) GetUserFromSession(ctx context.Context, key string) (authn.User, error) {
	if svr.Session == nil {
		return authn.Anonymous(), ErrSessionManagerNotInitialized
	}
	return GetSessionItem[authn.User](svr.Session, ctx, key)
}

func (svr *AppServer) AddUserToSession(ctx context.Context, user authn.User) (string, error) {
	if svr.Session == nil {
		return "", ErrSessionManagerNotInitialized
	}
	key := common.CreateRandString(16)

	for svr.Session.Exists(ctx, key) {
		key = common.CreateRandString(16)
	}

	err := svr.AddSession(ctx, key, user)
	if err != nil {
		return "", err
	}
	return key, nil
}

func (svr *AppServer) AddOrReplaceUserToSession(ctx context.Context, key string, user authn.User) (string, error) {
	svr.AddOrReplaceSession(ctx, key, user)
	return key, nil
}

func (svr *AppServer) RemoveSession(ctx context.Context, key string) error {
	if svr.Session == nil {
		return ErrSessionManagerNotInitialized
	}
	svr.Session.Remove(ctx, key)
	return nil
}

func (svr *AppServer) AddSession(ctx context.Context, key string, t any) error {
	if svr.Session == nil {
		return ErrSessionManagerNotInitialized
	}
	if svr.Session.Exists(ctx, key) {
		return ErrSessionKeyExists
	}
	svr.Session.Put(ctx, key, t)
	return nil
}

func (svr *AppServer) AddOrReplaceSession(ctx context.Context, key string, t any) error {
	if svr.Session == nil {
		return ErrSessionManagerNotInitialized
	}
	svr.Session.Put(ctx, key, t)
	return nil
}

func (svr *AppServer) ReplaceSession(ctx context.Context, key string, t any) error {
	if svr.Session == nil {
		return ErrSessionManagerNotInitialized
	}
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
