package disk

import (
	"io/ioutil"
	"strings"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
)

type CertificateStorage struct {
	directory string
}

func (c *CertificateStorage) GetCertificate(domains []string) (publicChain, privateKey []byte, err error) {
	publicChain, err = ioutil.ReadFile(c.directory + storage.GetCertificateChainFilename(domains))
	if err != nil {
		return nil, nil, err
	}

	privateKey, err = ioutil.ReadFile(c.directory + storage.GetPrivateKeyFilename(domains))
	if err != nil {
		return nil, nil, err

	}
	return publicChain, privateKey, err
}

func NewCertificateStorage(path string) *CertificateStorage {
	path = strings.TrimSuffix(path, "/") + "/"

	return &CertificateStorage{directory: path}
}
