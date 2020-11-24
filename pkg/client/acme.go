package client

import (
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme"
	astorage "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme/storage"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/storage"
)

type AcmeClientBuilder struct {
	accountStorage      *astorage.Account
	http01Port          string
	acmeEmail           string
	forLocalDevelopment bool
}

func NewAcmeBuilder(accountStorage storage.Storage) *AcmeClientBuilder {
	return &AcmeClientBuilder{
		accountStorage: &astorage.Account{Storage: accountStorage},
	}
}

func (a *AcmeClientBuilder) ForAccount(email string) *AcmeClientBuilder {
	a.acmeEmail = email

	return a
}
func (a *AcmeClientBuilder) WithHTTP01Challenge(port string) *AcmeClientBuilder {
	a.http01Port = port

	return a
}

func (a *AcmeClientBuilder) ForLocalDevelopment() *AcmeClientBuilder {
	a.forLocalDevelopment = true

	return a
}

// Build is going to validate and configure accounts, please note that this will spit errors on any failure
func (a *AcmeClientBuilder) Build() (*lego.Client, error) {
	account := acme.NewAccount(a.accountStorage, a.acmeEmail)
	if err := account.LoadFromStorage(); err != nil {
		account.SetNewPrivateKey()
	}

	config := lego.NewConfig(account)

	// @see deployments/development-linux/readme.md
	if a.forLocalDevelopment {
		config.CADirURL = "https://127.0.0.1:14000/dir"
		config.Certificate.KeyType = certcrypto.RSA2048
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	if a.http01Port != "" {
		// The HTTP01 spec enforces challenge traffic over 80/443. Envoys in the edge will proxy it to our custom port
		err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", a.http01Port))
		if err != nil {
			return nil, err
		}
	}

	if account.IsRegistered() {
		return client, nil
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, err
	}

	account.SaveRegistration(reg)
	if err := account.PersistToStorage(); err != nil {
		return nil, err
	}

	return client, nil
}
