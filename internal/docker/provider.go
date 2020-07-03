package docker

import (
	"context"
	swarmtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/golang/protobuf/ptypes"
	"time"
)

type SwarmProvider struct {
	dockerClient client.APIClient
}

func NewSwarmProvider() SwarmProvider {
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
		dockerClient: c,
	}
}

// ProvideXDS will convert swarm service definitions to our own Models
func (s SwarmProvider) ProvideXDS(ctx context.Context) (clusters []types.Resource, endpoints []types.Resource, err error) {
	services, err := s.dockerClient.ServiceList(ctx, swarmtypes.ServiceListOptions{})
	if err != nil {
		return
	}

	for _, service := range services {
		clusters = append(clusters, s.convertServiceToCluster(&service))
		endpoints = append(endpoints, s.convertServiceToEndpoint(&service))
	}

	return
}

func (s SwarmProvider) convertServiceToEndpoint(service *swarm.Service) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: "mytest",
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.SocketAddress_TCP,
									Address:  "localhost",
									PortSpecifier: &core.SocketAddress_PortValue{
										PortValue: 1337,
									},
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
		Name:                 "mytest",
		ConnectTimeout:       ptypes.DurationProto(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STATIC},
	}
}
