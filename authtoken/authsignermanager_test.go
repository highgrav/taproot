package authtoken

import (
	"fmt"
	"testing"
	"time"
)

func TestAuthSignerManager(t *testing.T) {
	token := "abcdef1234567890"
	asm := NewAuthSignerManager(100*time.Minute, 100*time.Minute, DefaultAuthSecretRotator)
	encToken := asm.NewEncryptedToken(token)
	fmt.Println("Encrypted ASM Token: " + encToken)
	authtoken, err := asm.DecryptToken(encToken)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if authtoken.Token != token {
		t.Error("Got wrong token, expected '" + token + "', got '" + authtoken.Token)
		return
	}
	fmt.Println("Encrypted Token Expires On: " + authtoken.ExpiresAt.String() + ", token: " + authtoken.Token)

	signToken, err := asm.NewSignedToken(token)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println("Signed ASM Token: " + signToken)
	authtoken, err = asm.VerifySignedToken(signToken)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if authtoken.Token != token {
		t.Error("Got wrong token, expected '" + token + "', got '" + authtoken.Token)
		return
	}
	fmt.Println("Signed Token Expires On: " + authtoken.ExpiresAt.String() + ", token: " + authtoken.Token)

}
