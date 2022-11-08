package awsmocker

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"os"
)

//go:embed cacert.pem
var caCert []byte

//go:embed cakey.pem
var caKey []byte

var caKeyPair tls.Certificate

func init() {
	keypair, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		panic("Error parsing internal CA " + err.Error())
	}
	caKeyPair = keypair

	cert, err := x509.ParseCertificate(caKeyPair.Certificate[0])
	if err != nil {
		panic("Error parsing internal CAcert " + err.Error())
	}

	caKeyPair.Leaf = cert
}

// Exports the PEM Bytes of the CA Certificate (if you need to use it)
func CACertPEM() []byte {
	return caCert
}

// Returns the parsed X509 Certificate
func CACert() *x509.Certificate {
	return caKeyPair.Leaf
}

func writeCABundle(filePath string) error {
	return os.WriteFile(filePath, caCert, 0o600)
}
