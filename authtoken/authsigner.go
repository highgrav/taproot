package authtoken

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/highgrav/taproot/common"
	"golang.org/x/crypto/blake2b"
	"io"
	"strconv"
	"strings"
	"time"
)

// AuthSigner handles basic encryption and signing duties for cookie and header security assertions
type AuthSigner struct {
	ID        string
	secret    []byte
	StartsAt  time.Time
	ExpiresAt time.Time
	Block     cipher.Block // AES block should be safe for concurrent access, unlike BLAKE2
}

// NewAuthSigner() creates an AuthSigner with a password and expiration date
func NewAuthSigner(expiresAfter time.Duration, id string, secret []byte) (AuthSigner, error) {
	asign := AuthSigner{
		ID:        id,
		secret:    secret,
		StartsAt:  time.Now(),
		ExpiresAt: time.Now().Add(expiresAfter),
	}
	b, err := aes.NewCipher(asign.secret)
	if err != nil {
		return asign, err
	}
	asign.Block = b
	return asign, nil
}

func (asign *AuthSigner) createTokenString(tokenValue string) string {
	expAt := asign.ExpiresAt.Unix()
	nonce := common.CreateRandString(10)
	return strconv.FormatInt(expAt, 10) + "||" + tokenValue + "||" + nonce
}

func (asign *AuthSigner) NewSignedToken(tokenValue string) (string, error) {
	str := asign.createTokenString(tokenValue)
	var bytes []byte = make([]byte, 0)
	bytes = append(bytes, []byte(str+"||")...)
	bytes = append(bytes, asign.secret...)
	h, err := blake2b.New256(asign.secret)
	if err != nil {
		return "", err
	}
	signature := h.Sum([]byte(str))
	resToken := asign.ID + "||" + base64.StdEncoding.EncodeToString(signature) + "||" + str
	return resToken, nil
}

func (asign *AuthSigner) VerifySignedToken(signToken string) (AuthToken, error) {
	elems := strings.SplitN(signToken, "||", 2)

	if len(elems) < 2 {
		return AuthToken{}, ErrMalformedToken
	}
	sigBytes, err := base64.StdEncoding.DecodeString(elems[0])
	if err != nil {
		return AuthToken{}, err
	}
	h, err := blake2b.New256(asign.secret)
	if err != nil {
		return AuthToken{}, err
	}
	signature := h.Sum([]byte(elems[1]))
	for x := 0; x < len(signature) && x < len(sigBytes); x++ {
		if signature[x] != sigBytes[x] {
			return AuthToken{}, ErrInvalidSignature
		}
	}

	return asign.tokenStringToAuthToken(elems[1])
}

func (asign *AuthSigner) NewEncryptedToken(tokenValue string) string {
	str := asign.createTokenString(tokenValue)
	cipherTxt := make([]byte, aes.BlockSize+len(str))
	iv := cipherTxt[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return ""
	}
	stream := cipher.NewCFBEncrypter(asign.Block, iv)
	stream.XORKeyStream(cipherTxt[aes.BlockSize:], []byte(str))

	return asign.ID + "||" + base64.RawStdEncoding.EncodeToString(cipherTxt)
}

func (asign *AuthSigner) DecryptToken(encToken string) (AuthToken, error) {
	encBytes, err := base64.RawStdEncoding.DecodeString(encToken)
	if err != nil {
		return AuthToken{}, err
	}

	if len(encToken) < aes.BlockSize {
		return AuthToken{}, errors.New("malformed ciphertext")
	}
	iv := encBytes[:aes.BlockSize]
	encToken = encToken[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(asign.Block, iv)
	stream.XORKeyStream(encBytes, encBytes)

	// EncBytes is padded to fit AES blocksize, so make sure we start the slice at aes.Blocksize when reading
	atoken, err := asign.tokenStringToAuthToken(string(encBytes[aes.BlockSize:]))
	if err != nil {
		return AuthToken{}, err
	}
	return atoken, nil
}

func (asign *AuthSigner) tokenStringToAuthToken(token string) (AuthToken, error) {
	atkn := AuthToken{}
	elems := strings.Split(token, "||")
	if len(elems) != 3 {
		return atkn, errors.New("malformed plaintext (" + string(token) + ")")
	}
	i, err := strconv.ParseInt(elems[0], 10, 64)
	if err != nil {
		return atkn, errors.New("malformed plaintext (" + string(elems[0]) + ")")
	}
	atkn.ExpiresAt = time.Unix(i, 0)
	atkn.Token = elems[1]
	atkn.Nonce = elems[2]
	return atkn, nil
}
