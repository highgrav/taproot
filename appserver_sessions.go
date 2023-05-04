package taproot

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/highgrav/taproot/v1/authn"
	"github.com/highgrav/taproot/v1/common"
	"github.com/highgrav/taproot/v1/logging"
	"github.com/highgrav/taproot/v1/session"
	"net/http"
)

var ErrSessionKeyExists = errors.New("session key already exists")
var ErrSessionKeyDoesNotExist = errors.New("session key does not exist")
var ErrSessionInvalidType = errors.New("attempt to cast to invalid type")
var ErrSessionManagerNotInitialized = errors.New("session manager not initialized")

func GetSessionItem[T any](ses *session.SessionManager, key string) (T, error) {
	var t T
	if !ses.Exists(key) {
		return t, ErrSessionKeyDoesNotExist
	}
	return session.GetFromStore[T](ses, key)
}

/*
AuthenticateUser() returns an authenticated user, if applicable, returning the user and error.
*/
func (svr *AppServer) AuthenticateUser(authReq authn.UserAuth) (authn.User, error) {
	user, err := svr.users.GetUserByAuth(authReq)
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "SESS", "error", "error getting user: "+err.Error())
		return authn.Anonymous(), err
	}
	return user, nil
}

/*
RegisterUser() authenticates a user and creates a new session for them, returning the user, session key, and error.
*/
func (svr *AppServer) RegisterUser(authReq authn.UserAuth) (authn.User, string, error) {
	if svr.Session == nil {
		return authn.Anonymous(), "", ErrSessionManagerNotInitialized
	}
	user, err := svr.users.GetUserByAuth(authReq)
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "SESS", "error", "error getting user: "+err.Error())
		return authn.Anonymous(), "", err
	}
	key, err := svr.AddUserToSession(user)
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "SESS", "error", "Error adding user to session: "+err.Error())
		return authn.Anonymous(), "", err
	}

	return user, key, nil
}

func (svr *AppServer) AddSessionHeader(w http.ResponseWriter, token string) error {
	var sessionToken string
	var err error
	if svr.Config.UseEncryptedSessionTokens {
		sessionToken = svr.SignatureMgr.NewEncryptedToken(token)
	} else {
		sessionToken, err = svr.SignatureMgr.NewSignedToken(token)
		if err != nil {
			return err
		}
	}
	w.Header().Set(SESSION_HEADER_KEY, sessionToken)
	return nil
}

func (svr *AppServer) AddSessionCookie(w http.ResponseWriter, token string) error {
	var sessionToken string
	var err error
	if svr.Config.UseEncryptedSessionTokens {
		sessionToken = svr.SignatureMgr.NewEncryptedToken(token)
	} else {
		sessionToken, err = svr.SignatureMgr.NewSignedToken(token)
		if err != nil {
			return err
		}
	}

	cookie := http.Cookie{
		Name:     SESSION_COOKIE_NAME,
		Value:    sessionToken,
		Path:     "/",
		MaxAge:   3600,
		Secure:   svr.Config.Sessions.CookieSecure,
		HttpOnly: svr.Config.Sessions.CookieHttpOnly,
		SameSite: svr.Config.Sessions.CookieSiteMode,
	}
	http.SetCookie(w, &cookie)
	return nil
}

func (svr *AppServer) GetUserFromSession(key string) (authn.User, error) {
	if svr.Session == nil {
		return authn.Anonymous(), ErrSessionManagerNotInitialized
	}
	u, err := GetSessionItem[[]byte](svr.Session, key)
	if err != nil {
		return authn.Anonymous(), err
	}
	var user authn.User
	err = json.Unmarshal(u, &user)
	if err != nil {
		return authn.Anonymous(), err
	}
	return user, nil
}

func (svr *AppServer) AddUserToSession(user authn.User) (string, error) {
	if svr.Session == nil {
		return "", ErrSessionManagerNotInitialized
	}
	key := svr.Config.Sessions.SessionKeyPrefix + common.CreateRandString(16)

	for svr.Session.Exists(key) {
		key = svr.Config.Sessions.SessionKeyPrefix + common.CreateRandString(16)
	}
	j, err := json.Marshal(user)
	if err != nil {
		return "", err
	}
	err = svr.AddSession(key, []byte(j))
	if err != nil {
		return "", err
	}
	return key, nil
}

func (svr *AppServer) AddOrReplaceUserToSession(ctx context.Context, key string, user authn.User) (string, error) {
	svr.AddOrReplaceSession(key, user)
	return key, nil
}

func (svr *AppServer) RemoveSession(key string) error {
	if svr.Session == nil {
		return ErrSessionManagerNotInitialized
	}
	svr.Session.Remove(key)
	return nil
}

func (svr *AppServer) AddSession(key string, t any) error {
	if svr.Session == nil {
		return ErrSessionManagerNotInitialized
	}
	if svr.Session.Exists(key) {
		return ErrSessionKeyExists
	}
	logging.LogToDeck(context.Background(), "info", "SESS", "info", "adding new session with ID "+key)
	err := svr.Session.Put(key, t)
	if err != nil {
		return err
	}
	return nil
}

func (svr *AppServer) AddOrReplaceSession(key string, t any) error {
	if svr.Session == nil {
		return ErrSessionManagerNotInitialized
	}
	logging.LogToDeck(context.Background(), "info", "SESS", "info", "adding or removing new session with ID "+key)
	svr.Session.Put(key, t)
	return nil
}

func (svr *AppServer) ReplaceSession(key string, t any) error {
	if svr.Session == nil {
		return ErrSessionManagerNotInitialized
	}
	if !svr.Session.Exists(key) {
		return ErrSessionKeyDoesNotExist
	}
	svr.Session.Put(key, t)
	return nil
}

func (srv *AppServer) handleSessionError(w http.ResponseWriter, r *http.Request, err error) {
	logging.LogToDeck(context.Background(), "error", "SESS", "error", err.Error())
	srv.ErrorResponse(w, r, 500, "session error encountered")
}
