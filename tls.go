package taproot

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func generateSelfSignedTlsCert() (*tls.Config, error) {
	t := &tls.Config{}
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// We need to generate a random serial number to prevent browsers from
	// complaining
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		return nil, err
	}
	certTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Issuer:       pkix.Name{},
		Subject: pkix.Name{
			Organization: []string{"Taproot Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derCertBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, publicKey(privateKey), privateKey)
	if err != nil {
		return nil, err
	}
	certBuf := &bytes.Buffer{}
	err = pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derCertBytes})
	if err != nil {
		return nil, err
	}

	keyBuf := &bytes.Buffer{}
	err = pem.Encode(keyBuf, pemBlockForKey(privateKey))
	if err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair(certBuf.Bytes(), keyBuf.Bytes())
	if err != nil {
		return nil, err
	}

	t.Certificates = []tls.Certificate{cert}
	return t, nil
}
