package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/nstapelbroek/envoy-swarm-control-plane/internal"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/snapshot"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/watcher"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	internalLogger "github.com/nstapelbroek/envoy-swarm-control-plane/internal/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/docker"
)

var (
	debug                   bool
	port                    uint
	ingressNetwork          string
	controlPlaneClusterName string
	letsEncryptEmail        string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Use debug logging")
	flag.UintVar(&port, "port", 9876, "Management server port")
	flag.StringVar(&ingressNetwork, "ingress-network", "", "The swarm network name or ID that all services share with the envoy instances")
	flag.StringVar(&controlPlaneClusterName, "control-plane-cluster-name", "control_plane", "Name of the cluster your envoy instances are contacting for ADS/SDS")
	flag.StringVar(&letsEncryptEmail, "lets-encrypt-email", "", "Enable letsEncrypt TLS  certificate issuing by providing a expiration notice email")
}

func main() {
	flag.Parse()
	internalLogger.BootLogger(debug)
	main := context.Background()

	snapshotStorage := cache.NewSnapshotCache(
		false,
		cache.IDHash{},
		internalLogger.Instance().WithFields(logger.Fields{"area": "snapshot-cache"}),
	)

	snsProvider := createSnsProvider()            // sns provider manages downstream TLS certificates
	adsProvider := createAdsProvider(snsProvider) // ads provider converts swarm services to clusters and listeners
	manager := snapshot.NewManager(
		adsProvider,
		snsProvider,
		snapshotStorage,
		internalLogger.Instance().WithFields(logger.Fields{"area": "snapshot-manager"}),
	)

	events := createEventProducers(main)
	go manager.Listen(events)
	go internal.RunXDSServer(main, snapshotStorage, port)

	waitForSignal(main)
}

// createEventProducers will create a single place where multiple async triggers can coordinate an update of our state
func createEventProducers(main context.Context) chan snapshot.UpdateReason {
	UpdateEvents := make(chan snapshot.UpdateReason)

	go watcher.ForSwarmEvent(internalLogger.Instance().WithFields(logger.Fields{"area": "watcher"})).Watch(main, UpdateEvents)
	//go watcher.ForCertificateExpiration(snsProvider,internalLogger.Instance().WithFields(logger.Fields{"area": "watcher"})).Watch(main, UpdateEvents)
	go watcher.CreateInitialStartupEvent(UpdateEvents)

	return UpdateEvents
}

func createAdsProvider(snsProvider provider.SDS) provider.ADS {
	return docker.NewSwarmProvider(
		ingressNetwork,
		snsProvider,
		internalLogger.Instance().WithFields(logger.Fields{"area": "ads-provider"}),
	)
}

func createSnsProvider() provider.SDS {
	return tls.NewCertificateSecretsProvider(
		controlPlaneClusterName,
		internalLogger.Instance().WithFields(logger.Fields{"area": "sns-provider"}),
	)
}

func waitForSignal(application context.Context) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	<-s
	internalLogger.Infof("SIGINT Received, shutting down...")
	application.Done()
}
