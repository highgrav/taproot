package session

import (
	"github.com/highgrav/taproot/v1/authn"
	"net/http"
	"time"
)

type SessionErrorFunc func(http.ResponseWriter, *http.Request, error)

type SessionManager struct {
	Lifetime  time.Duration
	Store     IStore
	ErrorFunc SessionErrorFunc
	Codec     ICodec
}

func NewSessionManager(store IStore) *SessionManager {
	return &SessionManager{
		Store: store,
		Codec: DefaultCodec{},
	}
}

func (ses *SessionManager) Exists(key string) bool {
	_, found, err := ses.Store.Find(key)
	if err != nil {
		return false
	}
	return found
}

func (ses *SessionManager) Put(key string, val any) error {
	encodedVal, err := ses.Codec.Encode(val)
	if err != nil {
		return err
	}
	err = ses.Store.Commit(key, encodedVal, time.Now().Add(30*time.Second))
	if err != nil {
		return err
	}
	return nil
}

func (ses *SessionManager) GetBytes(key string) ([]byte, error) {
	res, found, err := ses.Store.Find(key)
	if err != nil {
		return []byte{}, err
	}
	if !found {
		return []byte{}, ErrKeyNotInSession
	}
	return res, nil
}

func (ses *SessionManager) GetString(key string) (string, error) {
	str, err := GetFromStore[string](ses, key)
	if err != nil {
		return "", err
	}
	return str, nil
}

func (ses *SessionManager) GetUser(key string) (authn.User, error) {
	usr, err := GetFromStore[authn.User](ses, key)
	if err != nil {
		return authn.Anonymous(), err
	}
	return usr, nil
}

func (ses *SessionManager) Remove(key string) error {
	return ses.Store.Delete(key)
}

func GetFromStore[T any](ses *SessionManager, key string) (T, error) {
	var t T
	res, err := ses.GetBytes(key)
	if err != nil {
		return t, err
	}
	t, err = DecodeAs[T](ses.Codec, res)
	if err != nil {
		return t, err
	}
	return t, nil
}

func DecodeAs[T any](codec ICodec, encodedObj []byte) (T, error) {
	var t T
	_, err := codec.Decode(encodedObj, &t)
	if err != nil {
		return t, err
	}
	return t, nil
}
