package storage

import (
	"gotest.tools/assert"
	"path/filepath"
	"testing"
)

func TestFileNameGeneratorIsIdempotent(t *testing.T) {
	domains := []string{"something.com", "hello.co.uk", "www.hello.co.uk"}

	firstRun := getFilenameHash(domains)

	assert.Equal(t, firstRun, getFilenameHash(domains))
}

func TestPublicCertificateFilenameExtension(t *testing.T) {
	filename := GetPublicCertificateFilename([]string{"domain.com"})

	assert.Equal(t, filepath.Ext(filename), ".crt")
}

func TestPrivateKeyFilenameExtension(t *testing.T) {
	filename := GetPrivateKeyFilename([]string{"domain.com"})

	assert.Equal(t, filepath.Ext(filename), ".key")
}
