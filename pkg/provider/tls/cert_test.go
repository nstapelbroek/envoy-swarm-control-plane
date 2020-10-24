package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var malformedStructure = `-----BEGIN CERTIFICATE-----
f01bTJgaKI4N9Xyqodw5RcatOlBkiWb80nGPsOo2ri/oP8/KbpHWwMhLzgyXDDeq
ibtJfCb52he7l3x6KRThlGPPp1C0u7Vn8saT0/XR+fJZYsSvdzXScC2pzrTt7nzt
R20quVGea+q83eLqxRzlFqcPCiAtoMbOkqBQQe9NqR5BTod0mXMxHLmTsGfgvk/c
tbMN8uBGoRQsJHAoFF6PqWBA101d5+8ujGHfMsh2/zzuL7ylsRmjh//vjeuWYWSQ
8/vzvz5EiZ3EaNhZzCVc3H43FgQq1wgVJ1v1hSpFAIekN7nAyceg3M6aOBc6/CVY
d1DRCb/U1TUlA24y4i/Pgg==
-----END CERTIFICATE-----
`

func TestValidatePublicCertificateMalformedStructure(t *testing.T) {
	result := validatePublicCertificate([]byte(malformedStructure))

	assert.Error(t, result)
	assert.Contains(t, result.Error(), "asn1: structure error")
}

func TestValidatePublicCertificateNoPemEncoding(t *testing.T) {
	result := validatePublicCertificate([]byte("randomcertificatebyteswithoutaheader"))

	assert.EqualError(t, result, "no PEM block found")
}

func TestValidatePublicCertificateExpiredCertificate(t *testing.T) {
	cert := createCertificate(time.Now().AddDate(0, 0, -1))

	result := validatePublicCertificate(cert)

	assert.EqualError(t, result, "certificate expired")
}

func TestValidatePublicCertificateValidCertificate(t *testing.T) {
	cert := createCertificate(time.Now().Add(5 * time.Second))

	assert.NoError(t, validatePublicCertificate(cert))
}

func createCertificate(expiry time.Time) []byte {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test", Organization: []string{"Supercorp."}},
		NotBefore:    time.Now(),
		NotAfter:     expiry,
		KeyUsage:     x509.KeyUsageCertSign,
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		panic("could not generate cert for tests: " + err.Error())
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
}
