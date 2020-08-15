package internal

import (
	"context"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/discovery"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider"
)

type Discovery struct {
	xdsProvider   provider.XDS
	sdsProvider   provider.SDS
	snapshotCache cache.SnapshotCache
	logger        logger.Logger
}

func NewDiscovery(xds provider.XDS, sds provider.SDS, c cache.SnapshotCache, log logger.Logger) *Discovery {
	return &Discovery{
		xdsProvider:   xds,
		sdsProvider:   sds,
		snapshotCache: c,
		logger:        log,
	}
}

func (d *Discovery) Watch(updateChannel chan discovery.Reason) {
	for {
		reason := <-updateChannel
		if err := d.discoverSwarm(reason); err != nil {
			d.logger.Fatalf(err.Error()) // For now, we kill the application as I don't know in what cases we could recover
		}
	}
}

func (d *Discovery) discoverSwarm(reason discovery.Reason) error {
	d.logger.WithFields(logger.Fields{"reason": reason}).Debugf("Running service discovery")

	const discoveryTimeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), discoveryTimeout)
	defer cancel()

	clusters, listeners, err := d.xdsProvider.Provide(ctx)
	if err != nil {
		return err
	}

	secrets, err := d.sdsProvider.Provide(ctx)
	if err != nil {
		return err
	}

	return d.createSnapshot(clusters, listeners, secrets)
}

func (d *Discovery) createSnapshot(clusters, listeners, secrets []types.Resource) error {
	snapshot := cache.Snapshot{}
	version := time.Now().Format(time.RFC3339)
	snapshot.Resources[types.Listener] = cache.NewResources(version, listeners)
	snapshot.Resources[types.Cluster] = cache.NewResources(version, clusters)
	snapshot.Resources[types.Secret] = cache.NewResources(version, secrets)

	// todo this would be the point where we write it to all node ids?
	err := d.snapshotCache.SetSnapshot("test-id", snapshot)
	if err != nil {
		return err
	}

	d.logger.WithFields(logger.Fields{"cluster-count": len(clusters), "listener-count": len(listeners), "secrets-count": len(secrets)}).Debugf("Updated snapshot")

	return err
}
