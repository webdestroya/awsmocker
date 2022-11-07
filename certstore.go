package awsmocker

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math"
	"math/big"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	globalCertStore *CertStorage

	leafCertStart = time.Unix(time.Now().Unix()-2592000, 0) // 2592000  = 30 day
	leafCertEnd   = time.Unix(time.Now().Unix()+31536000, 0)
)

type CertStorage struct {
	certs sync.Map

	nextSerial atomic.Int64
	privateKey *rsa.PrivateKey
}

func (tcs *CertStorage) Fetch(hostname string) *tls.Certificate {

	icert, ok := tcs.certs.Load(hostname)
	if ok {
		cert := icert.(tls.Certificate)
		return &cert
	}

	return tcs.generateCert(hostname)
}

func (tcs *CertStorage) generateCert(hostname string) *tls.Certificate {

	caCert := CACert()

	template := x509.Certificate{
		SerialNumber: big.NewInt(tcs.nextSerial.Add(1)),
		Issuer:       caCert.Subject,
		Subject: pkix.Name{
			Country:            []string{"US"},
			Organization:       []string{"webdestroya"},
			OrganizationalUnit: []string{"awsmocker fake leaf"},
		},
		NotBefore: leafCertStart,
		NotAfter:  leafCertEnd,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if ip := net.ParseIP(hostname); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, hostname)
		template.Subject.CommonName = hostname
	}

	var derBytes []byte
	var err error
	if derBytes, err = x509.CreateCertificate(rand.Reader, &template, caCert, tcs.privateKey.Public(), caKeyPair.PrivateKey); err != nil {
		panic("could not generate cert")
	}
	newCert := tls.Certificate{
		Certificate: [][]byte{derBytes, caKeyPair.Certificate[0]},
		PrivateKey:  tcs.privateKey,
	}

	val, _ := tcs.certs.LoadOrStore(hostname, newCert)

	val2 := (val).(tls.Certificate)
	return &val2
}

func init() {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	globalCertStore = &CertStorage{
		certs:      sync.Map{},
		privateKey: privKey,
	}

	startSerial, _ := rand.Int(rand.Reader, big.NewInt(int64(math.Pow(2, 40))))
	globalCertStore.nextSerial.Store(startSerial.Int64())
}
