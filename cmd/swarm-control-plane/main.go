package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/discovery"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/producer"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal"
	internalLogger "github.com/nstapelbroek/envoy-swarm-control-plane/internal/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/docker"
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
	internalLogger.BootLogger(debug)
	main := context.Background()

	// serving snapshots to our proxies
	snapshotCache := cache.NewSnapshotCache(
		true,
		cache.IDHash{},
		internalLogger.Instance().WithFields(logger.Fields{"area": "snapshot-cache"}),
	)

	go internal.RunGRPCServer(main, snapshotCache, port)

	// Internals to produce new snapshots
	UpdateEvents := make(chan discovery.Reason)

	provider := docker.NewSwarmProvider(
		ingressNetwork,
		internalLogger.Instance().WithFields(logger.Fields{"area": "provider"}),
	)

	consumer := internal.NewDiscovery(
		provider,
		snapshotCache,
		internalLogger.Instance().WithFields(logger.Fields{"area": "discovery"}),
		nodeID,
	)

	sp := producer.NewSwarmEventProducer(
		provider,
		internalLogger.Instance().WithFields(logger.Fields{"area": "swarm-events"}),
	)

	go consumer.Watch(UpdateEvents)
	go sp.UpdateOnSwarmEvents(main, UpdateEvents)
	go producer.InitialStartup(UpdateEvents)

	waitForSignal(main)
}

func waitForSignal(application context.Context) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	<-s
	internalLogger.Infof("SIGINT Received, shutting down...")
	application.Done()
}
