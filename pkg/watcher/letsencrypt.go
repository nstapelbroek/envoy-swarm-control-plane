package watcher

import (
	"context"
	"time"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/snapshot"
)

// LetsEncrypt is a poor man's interval trigger to issue any missing certificates at LetsEncrypt.
// I tried hooking this up with gRPC server callbacks but that lead to a lot of coupling
// between several control plane components. This is probably good enough for now :)
type LetsEncrypt struct {
	integration *acme.Integration
	logger      logger.Logger
}

func ForNewCertificates(integration *acme.Integration, log logger.Logger) *LetsEncrypt {
	return &LetsEncrypt{
		integration: integration,
		logger:      log,
	}
}

func (l *LetsEncrypt) Start(ctx context.Context, dispatchChannel chan snapshot.UpdateReason) {
	const UpdateInterval = 60

	for {
		select {
		case <-time.After(UpdateInterval * time.Second):
			l.logger.Debugf("LetsEncrypt issue interval tick")
			if reloadRequired, _ := l.integration.IssueCertificates(); reloadRequired {
				dispatchChannel <- "new LetsEncrypt certificate"
			}
		case <-ctx.Done():
			l.logger.Debugf("Stopping certificate issuing")
			return
		}
	}
}
