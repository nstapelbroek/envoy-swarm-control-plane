package main

import (
	"context"
	"flag"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/docker"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/logger"
	"os"
	"os/signal"
	"syscall"
)

var (
	debug          bool
	port           uint
	pollInterval   uint
	nodeID         string
	ingressNetwork string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Use debug logging")
	flag.UintVar(&port, "port", 9876, "Management server port")
	flag.UintVar(&pollInterval, "interval", 15, "Poll interval")
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")
	flag.StringVar(&ingressNetwork, "ingress-network", "", "The network name or ID which connects services to the loadbalancer")
}

func main() {
	flag.Parse()
	logger.BootLogger(debug)

	snapshotCache := cache.NewSnapshotCache(
		true,
		cache.IDHash{},
		logger.Instance().WithFields(logger.Fields{"area": "snapshot-cache"}),
	)
	provider := docker.NewSwarmProvider(
		ingressNetwork,
		logger.Instance().WithFields(logger.Fields{"area": "provider"}),
	)

	mainContext := context.Background()
	go internal.RunSwarmServiceDiscovery(mainContext, provider, snapshotCache, nodeID)
	go internal.RunGRPCServer(mainContext, snapshotCache, port)

	waitForSignal(mainContext)
}

func waitForSignal(ctx context.Context) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-s:
		logger.Infof("SIGINT Received, shutting down...")
		ctx.Done()
		return
	}
}
