package disk

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
)

type CertificateStorage struct {
	directory string
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

	return publicChain, privateKey, err
}

func NewCertificateStorage(path string) *CertificateStorage {
	path = strings.TrimSuffix(path, "/")

	return &CertificateStorage{directory: path}
}
