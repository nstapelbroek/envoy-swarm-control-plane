package conversion

import (
	"time"

	"github.com/docker/docker/api/types/swarm"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
)

// convertService will convert swarm service definitions into validated envoy resources
func ServiceToCluster(service *swarm.Service, labels *ServiceLabel) (c *cluster.Cluster, err error) {
	e := convertServiceToEndpoint(service, labels)
	if err = e.Validate(); err != nil {
		return
	}

	c = convertServiceToCluster(service, e)
	if err = c.Validate(); err != nil {
		return
	}

	return
}

func convertServiceToEndpoint(service *swarm.Service, labels *ServiceLabel) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: service.Spec.Name,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol:      labels.Endpoint.Protocol,
									Address:       "tasks." + service.Spec.Name,
									PortSpecifier: &labels.Endpoint.Port,
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func convertServiceToCluster(service *swarm.Service, loadAssignment *endpoint.ClusterLoadAssignment) *cluster.Cluster {
	const UpstreamConnectTimeout = 2 * time.Second
	const DNSRefreshRate = 4 * time.Second // When updating services, swarms default delay is 5 seconds, setting this to 4 leaves us with a 1 drain time (worst case)
	const UpstreamTCPKeepaliveProbes = 3
	const UpstreamTCPKeepaliveTime = 3600
	const UpstreamTCPKeepaliveInterval = 60

	return &cluster.Cluster{
		Name:                 service.Spec.Name,
		ConnectTimeout:       ptypes.DurationProto(UpstreamConnectTimeout),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
		RespectDnsTtl:        false, // Default TTL is 600, which is too long in the case of scaling down
		DnsRefreshRate:       ptypes.DurationProto(DNSRefreshRate),
		LoadAssignment:       loadAssignment,
		UpstreamConnectionOptions: &cluster.UpstreamConnectionOptions{
			// Unsure if these values make sense, I lowered the linux defaults as I expect the network to be more reliable than the www
			TcpKeepalive: &core.TcpKeepalive{
				KeepaliveProbes:   &wrappers.UInt32Value{Value: uint32(UpstreamTCPKeepaliveProbes)},
				KeepaliveTime:     &wrappers.UInt32Value{Value: uint32(UpstreamTCPKeepaliveTime)},
				KeepaliveInterval: &wrappers.UInt32Value{Value: uint32(UpstreamTCPKeepaliveInterval)},
			},
		},
	}
}
