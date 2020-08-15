package main

import (
	"context"
	"flag"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls"
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
	debug            bool
	port             uint
	ingressNetwork   string
	letsEncryptEmail string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Use debug logging")
	flag.UintVar(&port, "port", 9876, "Management server port")
	flag.StringVar(&ingressNetwork, "ingress-network", "", "The swarm network name or ID that all services share with the envoy instances")
	flag.StringVar(&letsEncryptEmail, "lets-encrypt-email", "", "Enable letsEncrypt TLS  certificate issuing by providing a expiration notice email")
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

	leProvider := tls.NewLetsEncryptProvider(
		letsEncryptEmail,
		internalLogger.Instance().WithFields(logger.Fields{"area": "letsencrypt-provider"}),
	)

	swarmProvider := docker.NewSwarmProvider(
		ingressNetwork,
		leProvider,
		internalLogger.Instance().WithFields(logger.Fields{"area": "swarm-provider"}),
	)

	consumer := internal.NewDiscovery(
		swarmProvider,
		leProvider,
		snapshotCache,
		internalLogger.Instance().WithFields(logger.Fields{"area": "discovery"}),
	)

	sp := producer.NewSwarmEventProducer(
		swarmProvider,
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
