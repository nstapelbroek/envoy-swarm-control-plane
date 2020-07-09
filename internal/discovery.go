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
	if err := discoverSwarm(p, c, nodeId); err != nil {
		// Any error during initial is going to cause os.exit as it guarantees fast feedback for initial setup.
		logger.Fatalf(err.Error(), err)
	}

	logger.Infof("initial service discovery done.")
	ticker := time.NewTicker(30 * time.Second)

	select {
	case <-ticker.C:
		// todo would be really cool on the long term to replae the ticker with an event listener
		// This might work out for us as we plan to rely fully on the routing mesh vip
		if err := discoverSwarm(p, c, nodeId); err != nil {
			logger.Errorf(err.Error(), err)
		}
	case <-ctx.Done():
		ticker.Stop()
		return
	}
}

func discoverSwarm(p docker.SwarmProvider, c cache.SnapshotCache, nodeId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// We don't support some resources yet, nullify them
	var routes, listeners, runtimes []types.Resource
	endpoints, clusters, err := p.ProvideADS(ctx)
	if err != nil {
		return err
	}

	snapShot := cache.NewSnapshot("1.0", endpoints, clusters, routes, listeners, runtimes)
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
	}).Debugf("Updated snapshot from Swarm Discovery")
	return nil
}
