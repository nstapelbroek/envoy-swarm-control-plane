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
	resourceProvider provider.Resource
	tlsProvider      provider.TLS
	snapshotCache    cache.SnapshotCache
	logger           logger.Logger
	nodeID           string
}

func NewDiscovery(p provider.Resource, c cache.SnapshotCache, log logger.Logger, nodeID string) *Discovery {
	return &Discovery{
		resourceProvider: p,
		tlsProvider:      nil,
		snapshotCache:    c,
		logger:           log,
		nodeID:           nodeID,
	}
}

func (d Discovery) Watch(updateChannel chan discovery.Reason) {
	for {
		reason := <-updateChannel
		if err := d.discoverSwarm(reason); err != nil {
			d.logger.Fatalf(err.Error()) // For now, we kill the application as I don't know in what cases we could recover
		}
	}
}

func (d Discovery) discoverSwarm(reason discovery.Reason) error {
	d.logger.WithFields(logger.Fields{"reason": reason}).Debugf("Running service discovery")

	const discoveryTimeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), discoveryTimeout)
	defer cancel()

	// endpoints and routes are embedded in the clusters and listeners. Other resources are not yet supported
	var endpoints, routes, runtimes, listeners []types.Resource

	clusters, httpListener, err := d.resourceProvider.ProvideClustersAndListener(ctx)
	if err != nil {
		return err
	}

	if d.tlsProvider != nil {
		listeners = d.tlsProvider.UpgradeHttpListener(httpListener)
	} else {
		listeners = append(listeners, httpListener)
	}

	return d.createSnapshot(endpoints, clusters, routes, listeners, runtimes, err)
}

func (d Discovery) createSnapshot(endpoints []types.Resource, clusters []types.Resource, routes []types.Resource, listeners []types.Resource, runtimes []types.Resource, err error) error {
	// todo this would be the point where we write it to all node ids?
	currentTime := time.Now()
	snapShot := cache.NewSnapshot(currentTime.Format(time.RFC3339), endpoints, clusters, routes, listeners, runtimes)
	err = d.snapshotCache.SetSnapshot(d.nodeID, snapShot)
	if err != nil {
		return err
	}

	d.logger.WithFields(logger.Fields{
		"endpoint-count": len(endpoints),
		"cluster-count":  len(clusters),
		"route-count":    len(routes),
		"listener-count": len(listeners),
		"runtime-count":  len(runtimes),
	}).Debugf("Updated snapshot")

	return nil
}
