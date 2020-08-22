package storage

import (
	"crypto/sha1" // #nosec since I'm only using it generate a file name
	"encoding/hex"
	"sort"
	"strings"
)

// We'll use the PEM container format, try to keep constants here so i'll return to this opinionated place :)
const CertificateExtension = "crt"
const PrivateKeyExtension = "key"

func GetCertificateChainFilename(domains []string) string {
	return GetCertificateFilename(domains) + "." + CertificateExtension
}

func GetPrivateKeyFilename(domains []string) string {
	return GetCertificateFilename(domains) + "." + PrivateKeyExtension
}

// GetCertificateFilename will transform a long list of domains into a sha1 string
func GetCertificateFilename(domains []string) string {
	h := sha1.New() // #nosec
	sort.Strings(domains)

	_, err := h.Write([]byte(strings.Join(domains, "")))
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(h.Sum(nil))
}
