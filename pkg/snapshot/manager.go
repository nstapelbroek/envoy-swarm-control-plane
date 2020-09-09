package snapshot

import (
	"context"
	"time"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider"
)

type Manager struct {
	adsProvider   provider.ADS
	sdsProvider   provider.SDS
	leIntegration *acme.Integration
	snapshotCache cache.SnapshotCache
	logger        logger.Logger
}

func NewManager(ads provider.ADS, sds provider.SDS, le *acme.Integration, c cache.SnapshotCache, log logger.Logger) *Manager {
	return &Manager{
		adsProvider:   ads,
		sdsProvider:   sds,
		leIntegration: le,
		snapshotCache: c,
		logger:        log,
	}
}

func (d *Manager) Listen(updateChannel chan UpdateReason) {
	for {
		reason := <-updateChannel
		if err := d.runDiscovery(reason); err != nil {
			d.logger.Fatalf(err.Error()) // For now, we kill the application as I don't know in what cases we could recover
		}
	}
}

func (d *Manager) runDiscovery(reason UpdateReason) error {
	d.logger.WithFields(logger.Fields{"reason": reason}).Debugf("Running service discovery")

	const discoveryTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), discoveryTimeout)
	defer cancel()

	clusters, listeners, err := d.adsProvider.Provide(ctx)
	if err != nil {
		return err
	}

	secrets, err := d.sdsProvider.Provide(ctx)
	if err != nil {
		return err
	}

	return d.createSnapshot(clusters, listeners, secrets)
}

func (d *Manager) createSnapshot(clusters, listeners, secrets []types.Resource) error {
	snapshot := cache.Snapshot{}
	version := time.Now().Format(time.RFC3339)
	snapshot.Resources[types.Listener] = cache.NewResources(version, listeners)
	snapshot.Resources[types.Cluster] = cache.NewResources(version, clusters)
	snapshot.Resources[types.Secret] = cache.NewResources(version, secrets)
	err := snapshot.Consistent()
	if err != nil {
		return err
	}

	err = d.snapshotCache.SetSnapshot(staticHash, snapshot)
	if err != nil {
		return err
	}

	d.logger.WithFields(logger.Fields{"cluster-count": len(clusters), "listener-count": len(listeners), "secrets-count": len(secrets)}).Debugf("Updated snapshot")

	return err
}
