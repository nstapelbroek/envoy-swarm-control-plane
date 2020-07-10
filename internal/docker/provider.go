package docker

import (
	"context"
	"errors"
	swarmtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/golang/protobuf/ptypes"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/logger"
	"net"
	"time"
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

// ProvideADS will convert swarm service definitions to our own Models
func (s SwarmProvider) ProvideADS(ctx context.Context) (endpoints []types.Resource, clusters []types.Resource, err error) {
	// make sure we have up-to-date info about our ingress network
	ingress, err := s.getIngressNetwork(ctx)
	if err != nil {
		return
	}

	// although introspecting by network makes more sense, I prefer debug output why we skipped a specific service
	services, err := s.dockerClient.ServiceList(ctx, swarmtypes.ServiceListOptions{})
	if err != nil {
		return
	}

	for _, service := range services {
		log := s.logger.WithFields(logger.Fields{"swarm-service-id": service.ID})
		ip := getIngressServiceIP(&service, &ingress)
		if ip == nil {
			log.Debugf("skipped generating ADS for service because it's not connected to the ingress network")
			continue
		}

		// We need at least a port where we can route the traffic towards, the rest can be defaulted
		labels := ParseServiceLabels(service.Spec.Labels)
		if len(labels.endpoints) < 1 {
			log.Debugf("skipped generating ADS for service because it has no port labeled")
			continue
		}

		clusters = append(clusters, s.convertServiceToCluster(&service))
		endpoints = append(endpoints, s.convertServiceToEndpoint(service.Spec.Name, ip, &labels))
	}

	return
}

func getIngressServiceIP(service *swarm.Service, ingress *swarmtypes.NetworkResource) net.IP {
	for _, vip := range service.Endpoint.VirtualIPs {
		if vip.NetworkID == ingress.ID {
			ip, _, _ := net.ParseCIDR(vip.Addr)
			return ip
		}
	}

	return nil
}

func (s SwarmProvider) convertServiceToEndpoint(clusterName string, ingressIP net.IP, labels *ServiceLabels) *endpoint.ClusterLoadAssignment {
	endpoints := make([]*endpoint.LbEndpoint, len(labels.endpoints))
	for _, serviceEndpoint := range labels.endpoints {
		endpoints = append(endpoints, &endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: &core.Address{
						Address: &core.Address_SocketAddress{
							SocketAddress: &core.SocketAddress{
								Protocol:      serviceEndpoint.protocol,
								Address:       ingressIP.String(),
								PortSpecifier: &serviceEndpoint.port,
							},
						},
					},
				},
			},
		})
	}

	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints:   []*endpoint.LocalityLbEndpoints{{LbEndpoints: endpoints}},
	}
}

func (s SwarmProvider) convertServiceToCluster(service *swarm.Service) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 service.Spec.Name,
		ConnectTimeout:       ptypes.DurationProto(2 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STATIC},
	}
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
