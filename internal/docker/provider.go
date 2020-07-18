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
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal/logger"
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

	for _, service := range services {
		log := s.logger.WithFields(logger.Fields{"swarm-service-id": service.ID})
		// We need at least a port where we can route the traffic towards, the rest can be defaulted
		labels := ParseServiceLabels(service.Spec.Labels)
		if labels.endpoint.port.PortValue <= 0 {
			log.Debugf("skipped generating ADS for service because it has no port labeled")
			continue
		}

		// Prevent confusion by filtering out services that are not properly connected
		// DNS requests will return empty responses if a service is not connected to the shared ingress network
		if !inIngressNetwork(&service, &ingress) {
			log.Debugf("skipped generating ADS for service because it's not connected to the ingress network")
			continue
		}

		clusters = append(clusters, s.convertServiceToCluster(&service))
		endpoints = append(endpoints, s.convertServiceToEndpoint(&service, &labels))
		serviceRoute := s.convertServiceToRoute(&service, &labels)
		routes = append(routes, serviceRoute)
		listeners = append(listeners, s.convertServiceToListener(&service, &labels, serviceRoute))
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

func (s SwarmProvider) convertServiceToEndpoint(service *swarm.Service, labels *ServiceLabel) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: service.Spec.Name,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol:      labels.endpoint.protocol,
									Address:       "tasks." + service.Spec.Name,
									PortSpecifier: &labels.endpoint.port,
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func (s SwarmProvider) convertServiceToCluster(service *swarm.Service) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 service.Spec.Name,
		ConnectTimeout:       ptypes.DurationProto(2 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
		RespectDnsTtl:        false,                                 // Default TTL is 600, which is too long in the case of scaling
		DnsRefreshRate:       ptypes.DurationProto(5 * time.Second), // When scaling, swarm CLI awaits 5 seconds before marking the service converged
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

func (s SwarmProvider) convertServiceToRoute(service *swarm.Service, labels *ServiceLabel) *route.RouteConfiguration {
	// Assume that label parsing did not overwrite with an empty array
	primaryDomain := labels.route.domains[0]
	vhostName := service.Spec.Name
	if labels.route.domains[0] != "*" {
		vhostName = primaryDomain
	}

	return &route.RouteConfiguration{
		Name: service.Spec.Name + "_route",
		VirtualHosts: []*route.VirtualHost{{
			Name:    vhostName,
			Domains: labels.route.domains,
			Routes: []*route.Route{{
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: labels.route.path,
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: service.Spec.Name,
						},
					},
				},
			}},
		}},
	}
}

func (s SwarmProvider) convertServiceToListener(service *swarm.Service, labels *ServiceLabel, route *route.RouteConfiguration) *listener.Listener {
	manager := &hcm.HttpConnectionManager{
		CodecType:      hcm.HttpConnectionManager_AUTO,
		StatPrefix:     "http",
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{RouteConfig: route},
		HttpFilters: []*hcm.HttpFilter{{
			Name: "envoy.filters.http.router",
		}},
	}
	mngr, err := ptypes.MarshalAny(manager)
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name: service.Spec.Name + "_listener",
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0", // Default to all addresses since we don't know where our proxies are running
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: 80, // todo
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: mngr,
				},
			}},
		}},
	}
}
