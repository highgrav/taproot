package authtoken

import (
	"context"
	"github.com/highgrav/taproot/common"
	"github.com/highgrav/taproot/logging"
	"strings"
	"time"
)

type AuthSecretRotator func() (string, []byte, error)

func DefaultAuthSecretRotator() (string, []byte, error) {
	id := common.CreateRandString(8)
	secret := common.CreateRandBytes(32)
	return id, secret, nil
}

type AuthSignerManager struct {
	ExpiresAfter time.Duration
	GracePeriod  time.Duration

	Done                       chan bool
	CurrentSignatureExpiration time.Time
	currentSigner              *AuthSigner
	signers                    map[string]*AuthSigner
	ticker                     *time.Ticker
	rotator                    AuthSecretRotator
}

func NewAuthSignerManager(rotationTime time.Duration, gracePeriod time.Duration, rotator AuthSecretRotator) *AuthSignerManager {
	asm := &AuthSignerManager{
		ExpiresAfter:  rotationTime + gracePeriod,
		GracePeriod:   gracePeriod,
		Done:          make(chan bool),
		currentSigner: nil,
		signers:       make(map[string]*AuthSigner),
		ticker:        time.NewTicker(10 * time.Second), // TODO
		rotator:       rotator,
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

func (asm *AuthSignerManager) AddSigner() error {
	id, secret, err := asm.rotator()
	if err != nil {
		return err
	}
	asgn, err := NewAuthSigner(asm.ExpiresAfter, id, secret)
	if err != nil {
		logging.LogToDeck(context.Background(), "error", "AUTH", "error", "error adding signer: "+err.Error())
		return err
	}
	asm.signers[asgn.ID] = &asgn
	asm.currentSigner = &asgn
	asm.CurrentSignatureExpiration = asm.currentSigner.ExpiresAt
	return nil
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
