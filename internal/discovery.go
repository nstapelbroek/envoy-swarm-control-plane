package internal

import (
	"context"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/docker"
)

func RunSwarmServiceDiscovery(ctx context.Context, p docker.SwarmProvider, c cache.SnapshotCache, nodeId string) {
	err := discoverSwarm(p, c, nodeId)
	if err != nil {
		// todo complain upstream
		panic(err)
	}

	// integrate some sort of strategy here, choices:
	// 1. Interval Poll, just like Treaefik does
	// 2. Listen to events, as we are planning to route using DDNS or VIP we don't need to manage IP address state :)
}

func discoverSwarm(p docker.SwarmProvider, c cache.SnapshotCache, nodeId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2)
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
