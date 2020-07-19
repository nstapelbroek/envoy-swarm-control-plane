package docker

import (
	"context"
	"errors"
	swarmtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/logger"
)

type SwarmProvider struct {
	dockerClient   client.APIClient
	ingressNetwork string // the network name/id where our envoy communicates with services
	logger         logger.Logger
}

func NewSwarmProvider(ingressNetwork string, logger logger.Logger) SwarmProvider {
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
		logger:         logger,
	}
}

// ProvideADS will break down swarm service definitions into clusters, endpoints and vhosts. The vhosts are then aggregated into Envoy Routes and listeners
func (s SwarmProvider) ProvideADS(ctx context.Context) (
	endpoints []types.Resource,
	clusters []types.Resource,
	routes []types.Resource,
	listeners []types.Resource,
	err error) {
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

	var vhosts []*route.VirtualHost
	for _, service := range services {
		log := s.logger.WithFields(logger.Fields{"swarm-service-id": service.ID})

		// If any errors occur here, we'll just skip the service otherwise one config error can nullify the entire cluster
		cluster, endpoint, vhost, err := s.convertService(&service, &ingress)
		if err != nil {
			log.Warnf("skipped generating ADS for service because %s", err.Error())
			continue
		}

		clusters = append(clusters, cluster)
		endpoints = append(endpoints, endpoint)
		vhosts = append(vhosts, vhost)
	}

	r := s.configureRoutes(vhosts)
	routes = append(routes, r)
	listeners = append(listeners, s.configureHttpListener(r))

	return
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
