package disk

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
)

type CertificateStorage struct {
	directory string
}

func (c *CertificateStorage) GetStorageDirectory() string {
	return c.directory
}

func (c *CertificateStorage) PutCertificate(domain string, sans []string, publicChain, privateKey []byte) (err error) {
	fileName := storage.GetCertificateFilename(domain, sans)
	err = ioutil.WriteFile(
		fmt.Sprintf("%s/%s.%s", c.directory, fileName, storage.CertificateExtension),
		publicChain,
		os.FileMode(storage.CertificateFileMode),
	)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(
		fmt.Sprintf("%s/%s.%s", c.directory, fileName, storage.PrivateKeyExtension),
		privateKey,
		os.FileMode(storage.CertificateFileMode),
	)
}

func (c *CertificateStorage) GetCertificate(domain string, sans []string) (publicChain, privateKey []byte, err error) {
	fileName := storage.GetCertificateFilename(domain, sans)
	publicChain, err = ioutil.ReadFile(fmt.Sprintf("%s/%s.%s", c.directory, fileName, storage.CertificateExtension))
	if err != nil {
		return nil, nil, err
	}

	privateKey, err = ioutil.ReadFile(fmt.Sprintf("%s/%s.%s", c.directory, fileName, storage.PrivateKeyExtension))
	if err != nil {
		return nil, nil, err
	}

	// todo: in the future we need to think how we prevent expired certificates at this point

	return publicChain, privateKey, err
}

func NewCertificateStorage(path string) *CertificateStorage {
	path = strings.TrimSuffix(path, "/")

	return &CertificateStorage{directory: path}
}
