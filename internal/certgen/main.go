//go:build generate
// +build generate

package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

func main() {

	// keep the key size low. we don't care about security here, this is a local mock server
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:         "AWSMocker Root CA",
			Country:            []string{"US"},
			Organization:       []string{"webdestroya"},
			OrganizationalUnit: []string{"awsmocker"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(50, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		panic(err)
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	if err := os.WriteFile("./cacert.pem", caPEM.Bytes(), 0o644); err != nil {
		panic(fmt.Errorf("failed to write defaults file: %w", err))
	}

	if err := os.WriteFile("./cakey.pem", caPrivKeyPEM.Bytes(), 0o644); err != nil {
		panic(fmt.Errorf("failed to write defaults file: %w", err))
	}

}
