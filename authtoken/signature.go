package authtoken

import (
	"errors"
)

var (
	ErrInvalidSignature = errors.New("invalid signature, cannot decrypt")
	ErrMalformedToken   = errors.New("malformed token, cannot decrypt")
	ErrExpiredToken     = errors.New("token is expired, please log in again")
)
