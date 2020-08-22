package storage

import (
	"path/filepath"
	"testing"

	"gotest.tools/assert"
)

func TestFileNameGeneratorIsIdempotent(t *testing.T) {
	domains := []string{"something.com", "hello.co.uk", "www.hello.co.uk"}

	firstRun := GetCertificateFilename(domains)

	assert.Equal(t, firstRun, GetCertificateFilename(domains))
}

func TestPublicCertificateFilenameExtension(t *testing.T) {
	filename := GetCertificateChainFilename([]string{"domain.com"})

	assert.Equal(t, filepath.Ext(filename), ".crt")
}

func TestPrivateKeyFilenameExtension(t *testing.T) {
	filename := GetPrivateKeyFilename([]string{"domain.com"})

	assert.Equal(t, filepath.Ext(filename), ".key")
}
