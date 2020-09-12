package swarm

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types/swarm"

	swarmtypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/client"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/swarm/converting"
)

// ADSProvider will convert swarm service definitions and labels into cluster, listener and route configuration for Envoy
type ADSProvider struct {
	ingressNetwork  string // the network name/id where our envoy communicates with services
	dockerClient    docker.APIClient
	listenerBuilder *ListenerBuilder
	logger          logger.Logger
}

func NewADSProvider(ingressNetwork string, builder *ListenerBuilder, log logger.Logger) *ADSProvider {
	return &ADSProvider{
		dockerClient:    client.NewDockerClient(),
		listenerBuilder: builder,
		ingressNetwork:  ingressNetwork,
		logger:          log,
	}
}

func (s *ADSProvider) Provide(ctx context.Context) (clusters, listeners []types.Resource, err error) {
	clusters, vhosts, err := s.provideClustersAndVhosts(ctx)
	if err != nil {
		s.logger.Errorf("Failed creating clusters and vhost configurations")
		return nil, nil, err
	}

	listeners, err = s.listenerBuilder.ProvideListeners(vhosts)
	if err != nil {
		s.logger.Errorf("Failed converting the vhosts into a listener configuration")
		return nil, nil, err
	}

	return clusters, listeners, nil
}

// provideClustersAndVhosts will break down swarm service definitions into clusters, endpoints and vhosts
func (s *ADSProvider) provideClustersAndVhosts(ctx context.Context) (clusters []types.Resource, vhosts *converting.VhostCollection, err error) {
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

func (s *ADSProvider) getIngressNetwork(ctx context.Context) (network swarmtypes.NetworkResource, err error) {
	network, err = s.dockerClient.NetworkInspect(ctx, s.ingressNetwork, swarmtypes.NetworkInspectOptions{})
	if err != nil {
		return
	}

	if network.Scope != "swarm" {
		return network, errors.New("the provided ingress network is not scoped for the entire cluster (swarm)")
	}

	return
}

func inIngressNetwork(service *swarm.Service, ingress *swarmtypes.NetworkResource) bool {
	for _, vip := range service.Endpoint.VirtualIPs {
		if vip.NetworkID == ingress.ID {
			return true
		}
	}

	return false
}
