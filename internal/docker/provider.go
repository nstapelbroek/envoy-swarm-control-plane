package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
)

type SwarmProvider struct {
	dockerClient client.APIClient
	cdsConversions <-chan *swarm.Service
	edsConversions <-chan *swarm.Service
	rdsConversions <-chan *swarm.Service
	tdsConversions <-chan *swarm.Service
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
		cdsConversions: make(<-chan swarm.Service),
		edsConversions: make(<-chan swarm.Service),
		rdsConversions: make(<-chan swarm.Service),
		tdsConversions: make(<-chan swarm.Service),
	}
}

// ListServices will convert swarm service definitions to our own Models
func (s SwarmProvider) ListServices(ctx context.Context) ([]swarm.Service, error) {
	return s.dockerClient.ServiceList(ctx, types.ServiceListOptions{})
}

func (s SwarmProvider) convertServicesToClusters(services []swarm.Service) interface{} {
	for _, s := range services {
		cdsConversions <- &s
		edsConversions <- &s
		rdsConversions <- &s
		tdsConversions <- &s
		// todo sds for certificates
	}
	return envoy_config_cluster_v3.Cluster{
		TransportSocketMatches:          nil,
		Name:                            "",
		AltStatName:                     "",
		ClusterDiscoveryType:            nil,
		EdsClusterConfig:                nil,
		ConnectTimeout:                  nil,
		PerConnectionBufferLimitBytes:   nil,
		LbPolicy:                        0,
		LoadAssignment:                  nil,
		HealthChecks:                    nil,
		MaxRequestsPerConnection:        nil,
		CircuitBreakers:                 nil,
		UpstreamHttpProtocolOptions:     nil,
		CommonHttpProtocolOptions:       nil,
		HttpProtocolOptions:             nil,
		Http2ProtocolOptions:            nil,
		TypedExtensionProtocolOptions:                 nil,
		DnsRefreshRate:                                nil,
		DnsFailureRefreshRate:                         nil,
		RespectDnsTtl:                                 false,
		DnsLookupFamily:                               0,
		DnsResolvers:                                  nil,
		UseTcpForDnsLookups:                           false,
		OutlierDetection:                              nil,
		CleanupInterval:                               nil,
		UpstreamBindConfig:                            nil,
		LbSubsetConfig:                                nil,
		LbConfig:                                      nil,
		CommonLbConfig:                                nil,
		TransportSocket:                               nil,
		Metadata:                                      nil,
		ProtocolSelection:                             0,
		UpstreamConnectionOptions:                     nil,
		CloseConnectionsOnHostHealthFailure:           false,
		IgnoreHealthOnHostRemoval:                     false,
		Filters:                                       nil,
		LoadBalancingPolicy:                           nil,
		LrsServer:                                     nil,
		TrackTimeoutBudgets:                           false,
	}
}
