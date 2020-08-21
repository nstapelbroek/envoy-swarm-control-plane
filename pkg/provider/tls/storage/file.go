package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"sort"
	"strings"
)

// We'll use the PEM container format, try to keep constants here so i'll return to this opinionated place :)
const CertificateExtension = "crt"
const PrivateKeyExtension = "key"

func GetPublicCertificateFilename(domains []string) string {
	return getFilenameHash(domains) + "." + CertificateExtension
}

func GetPrivateKeyFilename(domains []string) string {
	return getFilenameHash(domains) + "." + PrivateKeyExtension
}

// getFilenameHash will transform a long list of domains into a sha1 string
func getFilenameHash(domains []string) string {
	h := sha1.New()
	sort.Strings(domains)

	h.Write([]byte(strings.Join(domains, "")))
	return hex.EncodeToString(h.Sum(nil))
}
