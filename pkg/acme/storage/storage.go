package storage

import (
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/storage"
)

type Account struct {
	storage.Storage
}

func (c *Account) LoadPrivateKeyAndRegistration(email string) (privateKey, registration []byte, err error) {
	privateKey, err = c.GetFile(privateKeyFileName(email))
	if err != nil {
		return nil, nil, err
	}

	registration, err = c.GetFile(registrationFileName(email))
	if err != nil {
		return nil, nil, err
	}

	return privateKey, registration, err
}

func (c *Account) SavePrivateKeyAndRegistration(email string, privateKey, registration []byte) error {
	if err := c.PutFile(privateKeyFileName(email), privateKey); err != nil {
		return err
	}

	return c.PutFile(registrationFileName(email), registration)
}
