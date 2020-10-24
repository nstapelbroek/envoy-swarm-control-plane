package storage

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
)

// We'll use .pem and .key files to store our certificate and private key, trying to play nice with x509 standards
// Note that this storage doesn't know anything about the encoding and merely exists to write bytes to disk
const CertificateExtension = "pem"
const PrivateKeyExtension = "key"

// getCertificateFilename contains the business logic for generating consistent certificate file names
func getCertificateFilename(primaryDomain string, sans []string) string {
	filename := strings.ToLower(primaryDomain)

	// Besides the human readable filename we need to add a hash
	// that causes a mismatch when SANs are added or removed from the array
	sortedDomains := make([]string, len(sans))
	_ = copy(sortedDomains, sans)
	sort.Strings(sortedDomains)

	sum := sha256.Sum256([]byte(strings.Join(sortedDomains, "")))
	hash := base64.StdEncoding.EncodeToString(sum[:])

	// The result should be unique enough to prevent a unintended collisions, 16 characters seems unique enough
	return strings.NewReplacer("/", "", "\\", "").Replace(fmt.Sprintf("%s-%s", filename, hash[:16]))
}
