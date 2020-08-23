package storage

import (
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestFileNameGeneratorIsIdempotent(t *testing.T) {
	domains := []string{"something.com", "hello.co.uk", "www.hello.co.uk"}

	firstRun := GetCertificateFilename(domains[0], domains)

	assert.Equal(t, firstRun, GetCertificateFilename(domains[0], domains))
}

func TestFileNameGeneratorContainsDomainName(t *testing.T) {
	domains := []string{"awesome.co.uk", "www.awesome.co.uk", "oldwebsite.com"}

	fileName := GetCertificateFilename(domains[0], domains)

	assert.Equal(t, strings.HasPrefix(fileName, "awesome.co.uk"), true)
}

func TestFileNameGeneratorIgnoresIllegalPathCharacters(t *testing.T) {
	domains := []string{"/etc/passwd"}
	domains2 := []string{"\\\\somehost\\directory"}

	fileName := GetCertificateFilename(domains[0], domains)
	fileName2 := GetCertificateFilename(domains2[0], domains2)

	assert.Check(t, !strings.Contains(fileName, "/"))
	assert.Check(t, !strings.Contains(fileName, "\\"))
	assert.Check(t, !strings.Contains(fileName2, "/"))
	assert.Check(t, !strings.Contains(fileName2, "\\"))
}

func TestFileNameGeneratorContainsHash(t *testing.T) {
	domains := []string{"awesome.co.uk", "www.awesome.co.uk", "oldwebsite.com"}

	fileName := GetCertificateFilename(domains[0], domains)

	assert.Equal(t, strings.HasSuffix(fileName, "z5ep8xrWar52XrUR"), true)
}

func TestFileNameGeneratorHashChangesWhenDomainsChange(t *testing.T) {
	domains := []string{"something.com", "hello.co.uk", "www.hello.co.uk"}

	firstRun := GetCertificateFilename(domains[0], domains)

	assert.Check(t, firstRun != GetCertificateFilename(domains[0], []string{"hello.co.uk"}))
}

func TestFileNameGeneratorCanHandleArrayIndexShifts(t *testing.T) {
	domains := []string{"something.com", "hello.co.uk", "www.hello.co.uk"}

	firstRun := GetCertificateFilename("something.com", domains)

	assert.Equal(t, firstRun, GetCertificateFilename("something.com", []string{"www.hello.co.uk", "hello.co.uk", "something.com"}))
}
