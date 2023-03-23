package authn

import (
	"context"
	"github.com/highgrav/taproot/v1/logging"
	"strings"
	"time"
)

type AuthSignerManager struct {
	ExpiresAfter time.Duration
	GracePeriod  time.Duration

	Done                       chan bool
	CurrentSignatureExpiration time.Time
	currentSigner              *AuthSigner
	signers                    map[string]*AuthSigner
	ticker                     *time.Ticker
}

func NewAuthSignerManager(rotationTime time.Duration, gracePeriod time.Duration) *AuthSignerManager {
	asm := &AuthSignerManager{
		ExpiresAfter:  rotationTime + gracePeriod,
		GracePeriod:   gracePeriod,
		Done:          make(chan bool),
		currentSigner: nil,
		signers:       make(map[string]*AuthSigner),
		ticker:        time.NewTicker(10 * time.Second), // TODO

	}
	asm.AddSigner()
	asm.CurrentSignatureExpiration = asm.currentSigner.ExpiresAt
	go asm.rotate()
	return asm
}

func (asm *AuthSignerManager) ListSignerKeys() []string {
	keys := make([]string, 0)
	for k, _ := range asm.signers {
		keys = append(keys, k)
	}
	return keys
}

func (asm *AuthSignerManager) rotate() {
	for {
		select {
		case <-asm.Done:
			return
		case t := <-asm.ticker.C:
			if t.After(asm.currentSigner.ExpiresAt.Add(time.Duration(-1) * asm.GracePeriod)) {
				currSig := asm.currentSigner.ID
				asm.AddSigner()
				logging.LogToDeck(context.Background(), "info", "AUTH", "info", "rotating session signer from "+currSig+" to "+asm.currentSigner.ID)
			}
			go asm.RemoveSigners()
		}
	}
}

func (asm *AuthSignerManager) AddSigner() {
	asgn, err := NewAuthSigner(asm.ExpiresAfter)
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "AUTH", "error", "error adding signer: "+err.Error())
		return
	}
	asm.signers[asgn.ID] = &asgn
	asm.currentSigner = &asgn
	asm.CurrentSignatureExpiration = asm.currentSigner.ExpiresAt
}

func (asm *AuthSignerManager) RemoveSigners() {
	toRem := make([]string, 0)
	for k, v := range asm.signers {
		if time.Now().After(v.ExpiresAt) {
			toRem = append(toRem, k)
		}
	}
	for _, r := range toRem {
		delete(asm.signers, r)
	}
}

func (asm *AuthSignerManager) NewSignedToken(valToEncrypt string) (string, error) {
	return asm.currentSigner.NewSignedToken(valToEncrypt)
}

func (asm *AuthSignerManager) VerifySignedToken(token string) (AuthToken, error) {
	elems := strings.SplitN(token, "||", 2)
	if len(elems) != 2 {
		return AuthToken{}, ErrMalformedToken
	}

	if s, ok := asm.signers[elems[0]]; ok {
		if time.Now().After(s.ExpiresAt) {
			return AuthToken{}, ErrExpiredToken
		}
		atok, err := s.VerifySignedToken(elems[1])
		return atok, err
	}
	return AuthToken{}, ErrExpiredToken
}

func (asm *AuthSignerManager) NewEncryptedToken(valToEncrypt string) string {
	return asm.currentSigner.NewEncryptedToken(valToEncrypt)
}

func (asm *AuthSignerManager) DecryptToken(token string) (AuthToken, error) {
	elems := strings.Split(token, "||")
	if len(elems) != 2 {
		return AuthToken{}, ErrMalformedToken
	}
	if s, ok := asm.signers[elems[0]]; ok {
		if time.Now().After(s.ExpiresAt) {
			return AuthToken{}, ErrExpiredToken
		}
		atok, err := s.DecryptToken(elems[1])
		return atok, err
	}
	return AuthToken{}, ErrExpiredToken
}
