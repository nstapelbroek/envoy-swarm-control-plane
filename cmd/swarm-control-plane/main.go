package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme"
	acmestorage "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme/storage"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/client"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls"
	tlsstorage "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/storage"

	"github.com/nstapelbroek/envoy-swarm-control-plane/internal"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/snapshot"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/watcher"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	internalLogger "github.com/nstapelbroek/envoy-swarm-control-plane/internal/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/docker"
)

var (
	debug            bool
	leTermsAccepted  bool
	xdsPort          uint
	lePort           uint
	ingressNetwork   string
	xdsClusterName   string
	leClusterName    string
	leEmail          string
	storagePath      string
	storageEndpoint  string
	storageBucket    string
	storageAccessKey string
	storageSecretKey string
)

func init() {
	// Required arguments with defaults shipped in the yaml
	flag.StringVar(&storagePath, "tls_storage-dir", "/etc/ssl/certs/le", "Local filesystem location where certificates are kept")
	flag.UintVar(&xdsPort, "xds-port", 9876, "The port where envoy instances can connect to for configuration updates")
	flag.UintVar(&lePort, "le-port", 8080, "The port where envoy will proxy lets encrypt HTTP-01 challenges towards")
	flag.StringVar(&ingressNetwork, "ingress-network", "edge-traffic", "The swarm network name or ID that all services share with the envoy instances")
	flag.StringVar(&xdsClusterName, "xds-cluster", "control_plane", "Name of the cluster your envoy instances are contacting for ADS/SDS")
	flag.StringVar(&leClusterName, "le-cluster", "control_plane_le", "Name of the cluster your envoy instances are contacting for ADS/SDS")

	// Required arguments for lets encrypt
	flag.StringVar(&leEmail, "le-email", "", "When registering for LetsEncrypt certificates this e-mail will be used for the account")
	flag.BoolVar(&leTermsAccepted, "le-accept-terms", false, "When registering for LetsEncrypt certificates this e-mail will be used for the account")

	// Optional arguments to store certificates in a object tls_storage
	flag.StringVar(&storageEndpoint, "tls_storage-endpoint", "certs3.amazonaws.com", "Host endpoint for the certs3 certificate tls_storage")
	flag.StringVar(&storageBucket, "tls_storage-bucket", "", "Bucket name of the certificate tls_storage")
	flag.StringVar(&storageAccessKey, "tls_storage-access-key", "", "Access key to authenticate at the certificate tls_storage")
	flag.StringVar(&storageSecretKey, "tls_storage-secret-key", "", "Secret key to authenticate at the certificate tls_storage")

	// Remainder flags
	flag.BoolVar(&debug, "debug", false, "Use debug logging")
}

func main() {
	flag.Parse()
	internalLogger.BootLogger(debug)
	main := context.Background()

	snapshotStorage := cache.NewSnapshotCache(
		false,
		snapshot.StaticHash{},
		internalLogger.Instance().WithFields(logger.Fields{"area": "snapshot-cache"}),
	)

	fileStorage := getStorage()
	snsProvider := createSnsProvider(tlsstorage.Certificate{Storage: fileStorage}) // sns provider manages downstream TLS certificates
	adsProvider := createAdsProvider(snsProvider)                                  // ads provider converts swarm services to clusters and listeners
	leIntegration := createLetsEncryptIntegration(
		acmestorage.User{Storage: fileStorage},
		tlsstorage.Certificate{Storage: fileStorage},
	)
	manager := snapshot.NewManager(
		adsProvider,
		snsProvider,
		leIntegration,
		snapshotStorage,
		internalLogger.Instance().WithFields(logger.Fields{"area": "snapshot-manager"}),
	)

	events := createEventProducers(main)
	go manager.Listen(events)
	go internal.RunXDSServer(main, snapshotStorage, xdsPort)
	waitForSignal(main)
}

func createLetsEncryptIntegration(userStorage acmestorage.User, certificateStorage tlsstorage.Certificate) *acme.Integration {
	if leTermsAccepted == false || leEmail == "" {
		return nil
	}

	return acme.NewIntegration(lePort, leEmail, userStorage, certificateStorage)
}

func getStorage() storage.Storage {
	disk := storage.NewDiskStorage(storagePath, internalLogger.Instance().WithFields(logger.Fields{"area": "disk"}))

	// return early when no s3 credentials are set
	if storageBucket == "" || storageAccessKey == "" || storageSecretKey == "" {
		return disk
	}

	minioClient, err := client.NewMinioClient(storageEndpoint, storageAccessKey, storageSecretKey)
	if err != nil {
		internalLogger.Instance().Fatalf(err.Error())
	}
	return storage.NewObjectStorage(minioClient, storageBucket, disk)
}

// createEventProducers will create a single place where multiple async triggers can coordinate an update of our state
func createEventProducers(main context.Context) chan snapshot.UpdateReason {
	UpdateEvents := make(chan snapshot.UpdateReason)
	log := internalLogger.Instance().WithFields(logger.Fields{"area": "watcher"})

	go watcher.ForSwarmEvent(log).Watch(main, UpdateEvents)
	//go watcher.ForNewTLSDomains(log).Watch(main, UpdateEvents)
	// go watcher.ForCertificateExpiration(snsProvider,internalLogger.Instance().WithFields(log.Fields{"area": "watcher"})).Watch(main, UpdateEvents)
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

func createSnsProvider(certificateStorage tlsstorage.Certificate) provider.SDS {
	return tls.NewCertificateSecretsProvider(
		xdsClusterName,
		certificateStorage,
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
