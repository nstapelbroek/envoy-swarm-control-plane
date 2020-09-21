package watcher

import (
	"context"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/snapshot"
)

type LetsEncrypt struct {
	integration     *acme.Integration
	snapshotStorage cache.ConfigWatcher
	logger          logger.Logger
}

func ForMissingCertificates(integration *acme.Integration, snapshotStorage cache.ConfigWatcher, log logger.Logger) *LetsEncrypt {
	return &LetsEncrypt{
		integration:     integration,
		snapshotStorage: snapshotStorage,
		logger:          log,
	}
}

func (s *LetsEncrypt) Start(ctx context.Context, dispatchChannel chan snapshot.UpdateReason) {
	responseChannel, _ := s.snapshotStorage.CreateWatch(cache.Request{TypeUrl: resource.ListenerType, Node: &core.Node{Id: "letsencrypt-startup-watcher"}})

	select {
	case _ = <-responseChannel:
		s.logger.Debugf("Snapshot storage has served configuration, (re) starting letEncrypt issuing")
		if reloadRequired, _ := s.integration.IssueCertificates(); reloadRequired {
			dispatchChannel <- "issuing LetsEncrypt certificates complete"
		}
	case <-ctx.Done():
		close(responseChannel)
	}
}
