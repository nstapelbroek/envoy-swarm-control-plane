package tls

import (
	"crypto/tls"
	"crypto/x509"
	"time"
)

// IsCertUsable is a place where we implement business logic to assure certificates are usable for envoy
func IsCertUsable(cert *tls.Certificate) bool {
	// assuming that index 0 is the leaf
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return false
	}

	return (time.Now()).Before(leaf.NotAfter)
}
