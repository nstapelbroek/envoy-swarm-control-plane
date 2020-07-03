package internal

import (
	"context"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/docker"
	"time"
)

// de provider abstracheert de docker specifics en provide XDS resources
// de provider wordt aangestuurd door een discovery class (manager, whatever) deze handeld intervals of signals af en schiet updated xds resources naar een caching struct
// een onderdeel dat de cache afhangt.

func RunSwarmServiceDiscovery(ctx context.Context, p docker.SwarmProvider, c cache.SnapshotCache, nodeId string) {
	if err := discoverSwarm(p, c, nodeId); err != nil {
		// Any error during initial is going to cause os.exit as it guarantees fast feedback for initial setup.
		Logger.Fatal(err)
	}

	Logger.Info("initial service discovery done.")
	ticker := time.NewTicker(30 * time.Second)

	select {
	case <-ticker.C:
		// todo would be really cool on the long term to replae the ticker with an event listener
		// This might work out for us as we plan to rely fully on the routing mesh vip
		if err := discoverSwarm(p, c, nodeId); err != nil {
			Logger.Error(err)
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
	endpoints, clusters, err := p.ProvideXDS(ctx)
	if err != nil {
		return err
	}

	snapShot := cache.NewSnapshot("1.0", endpoints, clusters, routes, listeners, runtimes)
	err = c.SetSnapshot(nodeId, snapShot)
	if err != nil {
		return err
	}

	Logger.With(
		"endpoints", len(endpoints),
		"clusters", len(clusters),
		"routes", len(routes),
		"listeners", len(listeners),
		"runtimes", len(runtimes),
	).Debug("Updated snapshot from Swarm Discovery")
	return nil
}
