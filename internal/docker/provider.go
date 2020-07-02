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

func (s SwarmProvider) convertServiceToEndpoint(service *swarm.Service) (endpoint *endpoint.Endpoint) {
	return
}

func (s SwarmProvider) convertServiceToCluster(service *swarm.Service) *cluster.Cluster {
	source := &core.ConfigSource{
		ResourceApiVersion: core.ApiVersion_V3,
		ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
			ApiConfigSource: &core.ApiConfigSource{
				ApiType:                   core.ApiConfigSource_GRPC,
				SetNodeOnFirstMessageOnly: true,
				GrpcServices: []*core.GrpcService{{
					TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "henk"},
					},
				}},
			},
		},
	}

	connectTimeout := 5 * time.Second

	return &cluster.Cluster{
		Name:                 "henk",
		ConnectTimeout:       ptypes.DurationProto(connectTimeout),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: source,
		},
	}
}
