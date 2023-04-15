package authn

import (
	"encoding/base64"
	"errors"
	"strings"
)

const (
	AUTH_BASIC          string = "basic"
	AUTH_BEARER         string = "bearer"
	AUTH_SESSION        string = "session"
	AUTH_FORM           string = "form"
	AUTH_TOTP           string = "totp"
	AUTH_MFACODE        string = "mfa"
	AUTH_VERIFICATION   string = "verification_code"
	AUTH_CODE           string = "recovery_code"
	AUTH_RESET_REQUEST  string = "reset_request"
	AUTH_PASSWORD_RESET string = "pwd_reset"
	AUTH_JWT            string = "jwt"
	AUTH_DIGEST         string = "digest"
	AUTH_OAUTH          string = "oauth"
	AUTH_HOBA           string = "hoba"
	AUTH_MUTUAL_TLS     string = "mutual_tls"
)

var (
	ErrUserNotFound               = errors.New("user not found")
	ErrUserNotAuthenticated       = errors.New("user was not authenticated")
	ErrUserRequiresAuthentication = errors.New("user must log in")
	ErrUserNotAuthorized          = errors.New("user was not authorized for authentication")
	ErrMalformedAuthHeader        = errors.New("malformed authorization header")
	ErrUnsupportedScheme          = errors.New("unsupported authorization header scheme")
	ErrAuthUnknownScheme          = errors.New("unknown authorization header scheme")
	ErrInvalidBasicCredentials    = errors.New("invalid basic credentials formatting")
)

type UserAuth struct {
	AuthType        string
	Realm           string
	Provider        string
	UserIdentifier  string
	PasswordOrToken string
	ResetToken      string
}

func ParseAuthHeader(hdr string) (UserAuth, error) {
	hdrElems := strings.Split(hdr, " ")
	if len(hdrElems) < 2 {
		return UserAuth{}, ErrMalformedAuthHeader
	}

	if hdrElems[0] != "Basic" && hdrElems[0] != "Bearer" && hdrElems[0] != "Digest" {
		return UserAuth{}, ErrAuthUnknownScheme
	}

	if hdrElems[0] == "Basic" {
		// try to decode
		sDec, err := base64.StdEncoding.DecodeString(hdrElems[1])
		if err != nil {
			return UserAuth{}, err
		}
		basicElems := strings.Split(string(sDec), ":")
		if len(basicElems) != 2 {
			return UserAuth{}, ErrInvalidBasicCredentials
		}
		ua := UserAuth{
			AuthType:        AUTH_BASIC,
			UserIdentifier:  basicElems[0],
			PasswordOrToken: basicElems[1],
		}

		return ua, nil
	}

	if hdrElems[0] == "Bearer" {
		ua := UserAuth{
			AuthType:        AUTH_BEARER,
			PasswordOrToken: hdrElems[1],
		}
		return ua, nil
	}

	return UserAuth{}, ErrUnsupportedScheme
}
