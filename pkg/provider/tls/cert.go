package tls

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"
)

// To make things less complex / MVP, I'm assuming:
// Certificates bytes are DER formatted in a PEM encoded file
// Private keys are RSA formatted in a PEM encoded (PKCS#8) file.
//
// Maybe we can add more ways in the future, but I'm hoping this standard suffices.

// validatePublicCertificate will return errors when the certificate is not properly formatted/encoded or has expired
func validatePublicCertificate(certBytes []byte) error {
	block, _ := pem.Decode(certBytes)
	if block == nil {
		return errors.New("no PEM block found")
	}

	if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
		return errors.New("PEM block is not a certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	if cert.NotAfter.Before(time.Now()) {
		return errors.New("certificate expired")
	}

	return nil
}
