package authn

import "time"

// AuthTokens hold basic security assertions for cookie and header security
type AuthToken struct {
	Token     string
	Nonce     string
	ExpiresAt time.Time
	Signature []byte
}
