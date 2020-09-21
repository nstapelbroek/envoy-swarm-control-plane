package acme

import (
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/go-acme/lego/v4/lego"
	acme "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme/storage"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	tls "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
	"sync"
)

type Integration struct {
	http01Port    uint
	http01Cluster string
	acmeEmail     string
	issueBacklog  [][]string
	mutex         sync.Mutex
	lego          *lego.Client
	certStorage   tls.Certificate
	logger        logger.Logger
}

func NewIntegration(userEmail, acmeClusterName string, acmeChallengePort uint, certStorage tls.Certificate, log logger.Logger) *Integration {
	return &Integration{
		http01Port:    acmeChallengePort,
		http01Cluster: acmeClusterName,
		issueBacklog:  [][]string{},
		acmeEmail:     userEmail,
		certStorage:   certStorage,
		logger:        log,
	}

	// We store the acme user keys in the same directory as the certificates for now
	u := acme.User{Storage: certStorage.Storage}

	return &Integration{http01Port: port, http01Route: r, userStorage: u, certStorage: certStorage, logger: log}
}

func (i *Integration) GetHTTP01Route() *route.Route {
	return i.http01Route
}

func (i *Integration) IssueCertificates() (reloadRequired bool, err error) {
	if len(i.issueBacklog) == 0 {
		return false, nil
	}

	return true, nil
}
