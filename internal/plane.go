package internal

import (
	"context"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

func StartPollingForChanges(ctx context.Context, p SwarmProvider, c cache.SnapshotCache) {
	//ticker := time.NewTicker(time.Second * time.Duration(pollInterval))

	//go func(p internal.SwarmProvider, ctx context.Context) {
	//	services, _ := p.Services(ctx)
	//	println(services)
	//}(p, ctx)

	services, _ := p.Services(ctx)
	println(services)

	sendDemoSnapshot(c)
}

func sendDemoSnapshot(c cache.SnapshotCache) {
	var clusters, endpoints, routes, listeners, runtimes []types.Resource
	snapshot := cache.NewSnapshot("1.0", endpoints, clusters, routes, listeners, runtimes)

	_ = c.SetSnapshot("some-node-id", snapshot)
}
