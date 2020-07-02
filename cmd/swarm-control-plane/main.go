package main

import (
	"context"
	"flag"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/docker"
	"os"
	"os/signal"
	"syscall"
)

var (
	debug        bool
	port         uint
	pollInterval uint
	nodeID       string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Use debug logging")
	flag.UintVar(&port, "port", 9876, "Management server port")
	flag.UintVar(&pollInterval, "interval", 15, "Poll interval")
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")
}

func main() {
	flag.Parse()
	internal.CreateLogger(debug)

	mainContext := context.Background()
	snapshotCache := cache.NewSnapshotCache(true, cache.IDHash{}, internal.Logger)
	provider := docker.NewSwarmProvider()

	go internal.RunSwarmServiceDiscovery(mainContext, provider, snapshotCache, nodeID)
	go internal.RunGRPCServer(mainContext, snapshotCache, port)

	waitForSignal(mainContext)
}

func waitForSignal(ctx context.Context) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-s:
		internal.Logger.Info("SIGINT Received, shutting down...")
		ctx.Done()
		return
	}
}
