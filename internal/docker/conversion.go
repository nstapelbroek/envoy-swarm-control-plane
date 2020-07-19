package docker

import (
	"errors"
	swarmtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	"time"
)

func (s SwarmProvider) configureRoutes(vhosts []*route.VirtualHost) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name:         "swarm_routes",
		VirtualHosts: vhosts,
	}
}

func (s SwarmProvider) configureHttpListener(route *route.RouteConfiguration) *listener.Listener {
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
		Name: "http_listener",
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

func (s SwarmProvider) convertServiceToRouteVHost(service *swarm.Service, labels *ServiceLabel) *route.VirtualHost {
	// Assume that label parsing did not overwrite with an empty array
	domain := labels.route.domains[0]
	name := service.Spec.Name
	if labels.route.domains[0] != "*" {
		name = domain
	}

	return &route.VirtualHost{
		Name:    name,
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
	}
}

// convertService will convert swarm service definitions into validated envoy resources
func (s SwarmProvider) convertService(service *swarm.Service, ingress *swarmtypes.NetworkResource) (cluster *cluster.Cluster, endpoint *endpoint.ClusterLoadAssignment, virtualHost *route.VirtualHost, err error) {
	// We need at least a port and domain where we can
	labels := ParseServiceLabels(service.Spec.Labels)
	if labels.endpoint.port.PortValue <= 0 {
		err = errors.New("there is no endpoint.port label specified")
		return
	}

	if len(labels.route.domains) == 0 {
		err = errors.New("there is no route.domains label specified")
		return
	}

	// Prevent confusion by filtering out services that are not properly connected
	// DNS requests will return empty responses if a service is not connected to the shared ingress network
	if !inIngressNetwork(service, ingress) {
		err = errors.New("the service is not connected to the ingress network")
		return
	}

	cluster = s.convertServiceToCluster(service)
	if err = cluster.Validate(); err != nil {
		return
	}

	endpoint = s.convertServiceToEndpoint(service, &labels)
	if err = endpoint.Validate(); err != nil {
		return
	}

	virtualHost = s.convertServiceToRouteVHost(service, &labels)
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
