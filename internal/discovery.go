package internal

import (
	"context"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/docker"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/logger"
)

func RunSwarmServiceDiscovery(ctx context.Context, p docker.SwarmProvider, c cache.SnapshotCache, nodeID string) {
	logger.Debugf("running initial swarm discovery")
	if err := discoverSwarm(p, c, nodeID); err != nil {
		// Any error during initial is going to cause os.exit as it guarantees fast feedback for initial setup.
		logger.Fatalf(err.Error())
	}

	tailContext, cancel := context.WithCancel(context.Background())
	logger.Debugf("starting event based discovery")
	go handleSwarmEvents(tailContext, p, c, nodeID)
	defer cancel()

	<-ctx.Done()
}

func handleSwarmEvents(ctx context.Context, p docker.SwarmProvider, c cache.SnapshotCache, nodeID string) {
	events, errorEvent := p.ListenForEvents(ctx)
	for {
		select {
		case <-events:
			logger.Debugf("received service event from docker, running discovery")
			if err := discoverSwarm(p, c, nodeID); err != nil {
				logger.Errorf(err.Error())
			}
		case err := <-errorEvent:
			logger.Errorf("received error while listening to swarm events: %s", err.Error())

			// Auto recover on errors @see github.com/docker/engine/client/events.go:19
			events, errorEvent = p.ListenForEvents(ctx)
		}
	}
}

func discoverSwarm(p docker.SwarmProvider, c cache.SnapshotCache, nodeID string) error {
	const dockerAPITimeout = 2 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), dockerAPITimeout)
	defer cancel()

	// endpoints and routes are embedded in the clusters and listeners. Other resources are not yet supported
	var endpoints, routes, runtimes []types.Resource
	clusters, listeners, err := p.ProvideClustersAndListeners(ctx)
	if err != nil {
		return err
	}

	currentTime := time.Now()
	snapShot := cache.NewSnapshot(currentTime.Format(time.RFC3339), endpoints, clusters, routes, listeners, runtimes)
	err = c.SetSnapshot(nodeID, snapShot)
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
