package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	streaming "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal"
	internalLogger "github.com/nstapelbroek/envoy-swarm-control-plane/internal/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/client"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/swarm"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls"
	tlsstorage "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/snapshot"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/storage"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/watcher"
)

var (
	debug            bool
	leTermsAccepted  bool
	acmeLocal        bool
	xdsPort          uint
	acmePort         string
	ingressNetwork   string
	xdsClusterName   string
	acmeClusterName  string
	acmeEmail        string
	storagePath      string
	storageEndpoint  string
	storageBucket    string
	storageAccessKey string
	storageSecretKey string
)

func init() {
	// Required arguments with defaults shipped in the yaml
	flag.StringVar(&storagePath, "storage-dir", "/etc/ssl/certs/le", "Local filesystem location where certificates are kept")
	flag.UintVar(&xdsPort, "xds-port", 9876, "The port where envoy instances can connect to for configuration updates")
	flag.StringVar(&acmePort, "acme-port", "8080", "The port where envoy will proxy lets encrypt HTTP-01 challenges towards")
	flag.StringVar(&ingressNetwork, "ingress-network", "edge-traffic", "The swarm network name or ID that all services share with the envoy instances")
	flag.StringVar(&xdsClusterName, "xds-cluster", "control_plane", "Name of the cluster your envoy instances are contacting for ADS/SDS")
	flag.StringVar(&acmeClusterName, "acme-cluster", "control_plane_acme", "Name of the cluster your envoy instances are proxying for ACME HTTP-01 challenges")

	// Required arguments for lets encrypt
	flag.StringVar(&acmeEmail, "acme-email", "", "When registering for LetsEncrypt certificates this e-mail will be used for the account")
	flag.BoolVar(&leTermsAccepted, "acme-accept-terms", false, "When registering for LetsEncrypt certificates this e-mail will be used for the account")

	// Optional arguments to store certificates in a object tls_storage
	flag.StringVar(&storageEndpoint, "storage-endpoint", "certs3.amazonaws.com", "Host endpoint for the certs3 certificate tls_storage")
	flag.StringVar(&storageBucket, "storage-bucket", "", "Bucket name of the certificate tls_storage")
	flag.StringVar(&storageAccessKey, "storage-access-key", "", "Access key to authenticate at the certificate tls_storage")
	flag.StringVar(&storageSecretKey, "storage-secret-key", "", "Secret key to authenticate at the certificate tls_storage")

	// Remainder flags
	flag.BoolVar(&debug, "debug", false, "Use debug logging")
	flag.BoolVar(&acmeLocal, "acme-local", false, "Use a local acme server setup for development")
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

	snsProvider, acmeIntegration := setupTLS()
	adsProvider := setupDiscovery(snsProvider, acmeIntegration)
	manager := snapshot.NewManager(
		adsProvider,
		snsProvider,
		snapshotStorage,
		internalLogger.Instance().WithFields(logger.Fields{"area": "snapshot-manager"}),
	)

	events := createWatchers(main, acmeIntegration)
	go manager.Listen(events)

	grpcHandler := streaming.NewServer(context.Background(), snapshotStorage, nil)
	go internal.RunXDSServer(main, grpcHandler, xdsPort)

	waitForSignal(main)
}

// createWatchers will boot all background watchers that can cause an state update in the control plane
func createWatchers(ctx context.Context, acmeIntegration *acme.Integration) chan snapshot.UpdateReason {
	UpdateEvents := make(chan snapshot.UpdateReason)
	log := internalLogger.Instance().WithFields(logger.Fields{"area": "watcher"})

	if acmeIntegration != nil {
		go watcher.ForLetsEncrypt(acmeIntegration, log).Start(ctx, UpdateEvents)
	}
	go watcher.ForSwarmEvent(log).Start(ctx, UpdateEvents)
	go watcher.CreateInitialStartupEvent(UpdateEvents)

	return UpdateEvents
}

// setupTLS will create an sds provider for sending tls certificates to clusters and an optional LetsEncrypt integration to issue new certificates
func setupTLS() (sdsProvider provider.SDS, acmeIntegration *acme.Integration) {
	fileStorage := getStorage()
	certificateStorage := &tlsstorage.Certificate{Storage: fileStorage}
	sdsProvider = tls.NewCertificateSecretsProvider(
		xdsClusterName,
		certificateStorage,
		internalLogger.Instance().WithFields(logger.Fields{"area": "sds-provider"}),
	)

	if !leTermsAccepted || acmeEmail == "" {
		return sdsProvider, acmeIntegration
	}

	// Due to complexity with registration and persisting state, we'll use a builder to split init logic
	acmeBuilder := client.NewAcmeBuilder(fileStorage).ForAccount(acmeEmail).WithHTTP01Challenge(acmePort)
	if acmeLocal {
		acmeBuilder.ForLocalDevelopment()
	}

	acmeClient, err := acmeBuilder.Build()
	acmeLogger := internalLogger.Instance().WithFields(logger.Fields{"area": "acme"})
	if err != nil {
		acmeLogger.Warnf("ACME integration disabled due to an initialisation error: %s", err.Error())
		return sdsProvider, acmeIntegration
	}

	return sdsProvider, acme.NewIntegration(
		acmeClient,
		acmeClusterName,
		certificateStorage,
		acmeLogger,
	)
}

// getStorage will configure the file with optional s3 extension
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

// setupDiscovery configures the discovery specifics that extracts clusters, endpoints, listeners and routes from swarm service's
func setupDiscovery(snsProvider provider.SDS, acmeIntegration *acme.Integration) provider.ADS {
	// Our Listener converter will contain logic to plug vhost into http or https listeners
	// while negotiating tls state at the SDS and LetEncrypt services
	listenerBuilder := swarm.NewListenerBuilder(
		snsProvider,
		acmeIntegration,
	)

	return swarm.NewADSProvider(
		ingressNetwork,
		listenerBuilder,
		internalLogger.Instance().WithFields(logger.Fields{"area": "ads-provider"}),
	)
}

func waitForSignal(application context.Context) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	<-s
	internalLogger.Infof("SIGINT Received, shutting down...")
	application.Done()
}
