package docker

import (
	"context"
	"errors"

	swarmtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/docker/converting"
)

type SwarmProvider struct {
	dockerClient   client.APIClient
	ingressNetwork string // the network name/id where our envoy communicates with services
	logger         logger.Logger
}

func NewSwarmProvider(ingressNetwork string, log logger.Logger) SwarmProvider {
	httpHeaders := map[string]string{
		"User-Agent": "Envoy Swarm Control Plane",
	}

	c, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithHTTPHeaders(httpHeaders),
	)
	if err != nil {
		panic(err)
	}

	return SwarmProvider{
		dockerClient:   c,
		ingressNetwork: ingressNetwork,
		logger:         log,
	}
}

func (s SwarmProvider) Events(ctx context.Context) (<-chan events.Message, <-chan error) {
	return s.dockerClient.Events(ctx, swarmtypes.EventsOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "type", Value: "service"}),
	})
}

// ProvideClustersAndListener will break down swarm service definitions into clusters and listeners internally those are composed of endpoints routes etc.
func (s SwarmProvider) ProvideClustersAndListener(ctx context.Context) (clusters []types.Resource, listener types.Resource, err error) {
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

	vhosts := converting.NewVhostCollection()
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

		err = vhosts.AddRoute(cluster.Name, labels)
		if err != nil {
			log.Warnf("skipped creating vhost for service because %s", err.Error())
			continue
		}

		clusters = append(clusters, cluster)
	}

	listener = vhosts.BuildListener()

	return clusters, listener, nil
}

func (s SwarmProvider) getIngressNetwork(ctx context.Context) (network swarmtypes.NetworkResource, err error) {
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
