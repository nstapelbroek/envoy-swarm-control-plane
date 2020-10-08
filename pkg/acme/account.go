package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/registration"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme/storage"
)

type Account struct {
	storage      *storage.Account
	email        string
	registration *registration.Resource
	privateKey   crypto.PrivateKey
}

func NewAccount(store *storage.Account, email string) *Account {
	return &Account{
		email:   email,
		storage: store,
	}
}

func (a *Account) GetEmail() string {
	return a.email
}

func (a *Account) GetRegistration() *registration.Resource {
	return a.registration
}

func (a *Account) GetPrivateKey() crypto.PrivateKey {
	return a.privateKey
}

func (a *Account) PersistToStorage() error {
	keyBytes := pem.EncodeToMemory(certcrypto.PEMBlock(a.privateKey))
	registrationBytes, err := json.MarshalIndent(a.registration, "", "\t")
	if err != nil {
		return err
	}

	return a.storage.SavePrivateKeyAndRegistration(a.email, keyBytes, registrationBytes)
}

func (a *Account) LoadFromStorage() error {
	privateBits, registrationBytes, err := a.storage.LoadPrivateKeyAndRegistration(a.email)
	if err != nil {
		return err
	}

	privateKey, err := a.decodePrivateKey(privateBits)
	if err != nil {
		return err
	}

	var reg registration.Resource
	a.privateKey = privateKey
	err = json.Unmarshal(registrationBytes, &reg)
	if err != nil {
		return err
	}

	a.registration = &reg
	return err
}

// Code is from go-acme/lego/v4@v4.0.1/cmd/accounts_storage.go:209
func (a *Account) decodePrivateKey(bytes []byte) (crypto.PrivateKey, error) {
	keyBlock, _ := pem.Decode(bytes)

	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(keyBlock.Bytes)
	}

	return nil, errors.New("unknown private privateKey type")
}

func (a *Account) SetNewPrivateKey() {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	a.privateKey = privateKey
}

func (a *Account) IsRegistered() bool {
	if a.registration == nil {
		return false
	}

	return a.registration.Body.Status == "valid"
}

func (a *Account) SaveRegistration(reg *registration.Resource) {
	a.registration = reg
}
