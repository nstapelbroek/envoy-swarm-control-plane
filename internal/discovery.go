package internal

import (
	"context"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/docker"
	"time"
)

func RunSwarmServiceDiscovery(ctx context.Context, p docker.SwarmProvider, c cache.SnapshotCache, nodeId string) {
	err := discoverSwarm(p, c, nodeId)
	if err != nil {
		// Any error during initial is going to cause os.exit for now :)
		Logger.Fatal(err)
	}

	Logger.Info("initial service discovery done.")
	select {
	case <-ctx.Done():
		return
		// integrate some sort of update strategy here, choices:
		// 1. Polling with a regular interval just like Treafik does
		// 2. Listening to docker events. This might work out for us as we plan to rely fully on the routing mesh vip
	}
}

func discoverSwarm(p docker.SwarmProvider, c cache.SnapshotCache, nodeId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	dockerServices, err := p.ListServices(ctx)
	if err != nil {
		return err
	}

	snapShot := createSnapshot(dockerServices)
	err = c.SetSnapshot(nodeId, snapShot)
	if err != nil {
		return err
	}

	return nil
}

func createSnapshot(services interface{}) cache.Snapshot {
	var clusters, endpoints, routes, listeners, runtimes []types.Resource
	snapshot := cache.NewSnapshot("1.0", endpoints, clusters, routes, listeners, runtimes)

	return snapshot
}
