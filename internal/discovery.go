package internal

import (
	"context"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/docker"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/logger"
	"time"
)

func RunSwarmServiceDiscovery(ctx context.Context, p docker.SwarmProvider, c cache.SnapshotCache, nodeId string) {
	logger.Debugf("starting swarm discovery")
	if err := discoverSwarm(p, c, nodeId); err != nil {
		// Any error during initial is going to cause os.exit as it guarantees fast feedback for initial setup.
		logger.Fatalf(err.Error())
	}

	pollContext, cancel := context.WithCancel(context.Background())
	swarmServiceChange := p.ListenToServiceChanges(pollContext)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			cancel()
			return
		case <-swarmServiceChange:
			logger.Infof("a swarm service event triggered a new swarm service discovery")
			if err := discoverSwarm(p, c, nodeId); err != nil {
				logger.Errorf(err.Error())
			}
		}
	}
}

func discoverSwarm(p docker.SwarmProvider, c cache.SnapshotCache, nodeId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// endpoints and routes are embedded in the clusters and listeners. Other resources are not yet supported
	var endpoints, routes, runtimes []types.Resource
	clusters, listeners, err := p.ProvideClustersAndListeners(ctx)
	if err != nil {
		return err
	}

	currentTime := time.Now()
	snapShot := cache.NewSnapshot(currentTime.Format(time.RFC3339), endpoints, clusters, routes, listeners, runtimes)
	err = c.SetSnapshot(nodeId, snapShot)
	if err != nil {
		return err
	}

	logger.WithFields(logger.Fields{
		"endpoint-count": len(endpoints),
		"cluster-count":  len(clusters),
		"route-count":    len(routes),
		"listener-count": len(listeners),
		"runtime-count":  len(runtimes),
	}).Debugf("Updated snapshot")
	return nil
}
