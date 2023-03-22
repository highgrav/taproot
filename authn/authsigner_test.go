package authn

import (
	"fmt"
	"testing"
	"time"
)

func TestGenString(t *testing.T) {
	asign, err := NewAuthSigner(100 * time.Minute)
	if err != nil {
		t.Error(err.Error())
	}
	str := asign.createTokenString("abc1234")
	fmt.Println("Token string: " + str)

	signedTok, err := asign.NewSignedToken(str)
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Println("Signed token: " + signedTok)

	encTok := asign.NewEncryptedToken(str)
	fmt.Println("Enc token: " + encTok)
}
