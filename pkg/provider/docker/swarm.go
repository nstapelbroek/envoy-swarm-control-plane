package docker

import (
	"context"
	"errors"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/client"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider"

	swarmtypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/docker/converting"
)

type SwarmProvider struct {
	ingressNetwork string // the network name/id where our envoy communicates with services
	dockerClient   docker.APIClient
	sdsProvider    provider.SDS
	acme           *acme.Integration
	logger         logger.Logger
}

func NewSwarmProvider(ingressNetwork string, sdsProvider provider.SDS, log logger.Logger, integration *acme.Integration) *SwarmProvider {
	return &SwarmProvider{
		dockerClient:   client.NewDockerClient(),
		sdsProvider:    sdsProvider,
		ingressNetwork: ingressNetwork,
		acme:           integration,
		logger:         log,
	}
}

// ProvideClustersAndListener will break down swarm service definitions into clusters and listeners internally those are composed of endpoints routes etc.
func (s *SwarmProvider) Provide(ctx context.Context) (clusters, listeners []types.Resource, err error) {
	clusters, vhosts, err := s.provideClustersAndVhosts(ctx)
	if err != nil {
		return nil, nil, err
	}

	if s.sdsProvider == nil {
		listeners = append(listeners, mapVhostsToHTTPListener(vhosts))
		return clusters, listeners, nil
	}

	listeners = mapVhostsToHTTPSListeners(vhosts, s.sdsProvider)

	return clusters, listeners, nil
}

func (s *SwarmProvider) provideClustersAndVhosts(ctx context.Context) (clusters []types.Resource, vhosts *converting.VhostCollection, err error) {
	// Make sure we have up-to-date info about our ingress network
	ingress, err := s.getIngressNetwork(ctx)
	if err != nil {
		return
	}

	// Although introspecting by network makes more sense, I prefer debug output why we skipped a specific service
	services, err := s.dockerClient.ServiceList(ctx, swarmtypes.ServiceListOptions{})
	if err != nil {
		return
	}

	vhosts = converting.NewVhostCollection()
	for i := range services {
		service := &services[i]
		log := s.logger.WithFields(logger.Fields{"swarm-service-name": service.Spec.Name})

		labels := converting.ParseServiceLabels(service.Spec.Labels)
		if err = labels.Validate(); err != nil {
			log.Debugf("skipping service because labels are invalid: %s", err.Error())
			continue
		}

		// Prevent confusion by filtering out services that are not properly connected
		// DNS requests will return empty responses if a service is not connected to the shared ingress network
		if !inIngressNetwork(service, &ingress) {
			log.Warnf("service is not connected to the ingress network, stopping processing")
			continue
		}

		cluster, err := converting.ServiceToCluster(service, labels)
		if err != nil {
			log.Warnf("skipped generating CDS for service because %s", err.Error())
			continue
		}

		err = vhosts.AddService(cluster.Name, labels)
		if err != nil {
			log.Warnf("skipped creating vhost for service because %s", err.Error())
			continue
		}

		clusters = append(clusters, cluster)
	}

	return clusters, vhosts, nil
}

func (s *SwarmProvider) getIngressNetwork(ctx context.Context) (network swarmtypes.NetworkResource, err error) {
	network, err = s.dockerClient.NetworkInspect(ctx, s.ingressNetwork, swarmtypes.NetworkInspectOptions{})
	if err != nil {
		return
	}

	if network.Scope != "swarm" {
		return network, errors.New("the provided ingress network is not scoped for the entire cluster (swarm)")
	}

	return
}
