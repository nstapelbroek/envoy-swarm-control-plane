package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidatePublicCertificateExpiredCertificate(t *testing.T) {
	cert := createCertificate(time.Now().AddDate(0, 0, -1))

	assert.False(t, IsCertUsable(&cert))
}

func TestValidatePublicCertificateValidCertificate(t *testing.T) {
	cert := createCertificate(time.Now().Add(5 * time.Second))

	assert.True(t, IsCertUsable(&cert))
}

func createCertificate(expiry time.Time) tls.Certificate {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test", Organization: []string{"Supercorp."}},
		NotBefore:    time.Now(),
		NotAfter:     expiry,
		KeyUsage:     x509.KeyUsageCertSign,
	}

	bytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		panic("could not generate bytes for tests: " + err.Error())
	}

	return tls.Certificate{Certificate: [][]byte{bytes}}
}
