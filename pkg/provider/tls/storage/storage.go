package storage

import (
	"fmt"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/storage"
)

type Certificate struct {
	storage.Storage
}

func (c *Certificate) PutCertificate(domain string, sans []string, publicChain, privateKey []byte) (err error) {
	fileName := getCertificateFilename(domain, sans)
	err = c.PutFile(fmt.Sprintf("%s.%s", fileName, CertificateExtension), publicChain)
	if err != nil {
		return err
	}

	return c.PutFile(fmt.Sprintf("%s.%s", fileName, PrivateKeyExtension), privateKey)
}

func (c *Certificate) GetCertificate(domain string, sans []string) (publicChain, privateKey []byte, err error) {
	fileName := getCertificateFilename(domain, sans)
	publicChain, err = c.GetFile(fmt.Sprintf("%s.%s", fileName, CertificateExtension))
	if err != nil {
		return nil, nil, err
	}

	privateKey, err = c.GetFile(fmt.Sprintf("%s.%s", fileName, PrivateKeyExtension))
	if err != nil {
		return nil, nil, err
	}

	return publicChain, privateKey, err
}
