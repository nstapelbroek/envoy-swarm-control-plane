package converting

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/wrapperspb"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

type FilterChainBuilder struct {
	name                 string
	configureTLS         bool
	sniServerNames       []string
	sdsCertificateConfig *auth.SdsSecretConfig
	vhosts               []*route.VirtualHost
}

func NewFilterChainBuilder(name string) *FilterChainBuilder {
	return &FilterChainBuilder{
		name:         name,
		configureTLS: false,
		vhosts:       []*route.VirtualHost{},
	}
}

func (b *FilterChainBuilder) EnableTLS(serverNames []string, sdsConfig *auth.SdsSecretConfig) *FilterChainBuilder {
	b.configureTLS = true
	b.sniServerNames = serverNames
	b.sdsCertificateConfig = sdsConfig

	return b
}

func (b *FilterChainBuilder) ForVhost(vhost *route.VirtualHost) *FilterChainBuilder {
	b.vhosts = append(b.vhosts, vhost)

	return b
}

func (b *FilterChainBuilder) Build() *listener.FilterChain {
	filterChain := listener.FilterChain{Name: b.name}
	if len(b.vhosts) > 0 {
		filterChain.Filters = append(filterChain.Filters, b.buildHTTPFilterForVhosts())
	}

	if b.configureTLS {
		filterChain.FilterChainMatch = &listener.FilterChainMatch{
			ServerNames: b.sniServerNames,
		}

		filterChain.TransportSocket = b.buildDownstreamTransportSocket()
	}

	return &filterChain
}

func (b *FilterChainBuilder) buildHTTPFilterForVhosts() *listener.Filter {
	const ServerName = "envoy-on-swarm/0.2"
	const RequestTimeout = 5 * time.Minute
	const MaxConcurrentHTTP2Streams = 100
	const InitialDownstreamHTTP2StreamWindowSize = 65536       // 64 KiB
	const InitialDownstreamHTTP2ConnectionWindowSize = 1048576 // 1 MiB

	routeType := "http"
	if b.configureTLS {
		routeType = "https"
	}

	routerConfig, _ := anypb.New(&router.Router{})
	routes := &route.RouteConfiguration{
		Name:         fmt.Sprintf("%s_%s_routes", b.name, routeType),
		VirtualHosts: b.vhosts,
	}
	conManager := &hcm.HttpConnectionManager{
		ServerName:       ServerName,
		CodecType:        hcm.HttpConnectionManager_AUTO,
		StatPrefix:       b.name,
		UseRemoteAddress: &wrapperspb.BoolValue{Value: true},
		RouteSpecifier:   &hcm.HttpConnectionManager_RouteConfig{RouteConfig: routes},
		HttpFilters: []*hcm.HttpFilter{{
			Name:       wellknown.Router,
			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
		}},
		Http2ProtocolOptions: &core.Http2ProtocolOptions{
			MaxConcurrentStreams:        &wrapperspb.UInt32Value{Value: uint32(MaxConcurrentHTTP2Streams)},
			InitialStreamWindowSize:     &wrapperspb.UInt32Value{Value: uint32(InitialDownstreamHTTP2StreamWindowSize)},
			InitialConnectionWindowSize: &wrapperspb.UInt32Value{Value: uint32(InitialDownstreamHTTP2ConnectionWindowSize)},
		},
		StreamIdleTimeout: durationpb.New(RequestTimeout),
		RequestTimeout:    durationpb.New(RequestTimeout),
	}
	managerConfig, err := anypb.New(conManager)
	if err != nil {
		panic(err)
	}

	return &listener.Filter{
		Name: wellknown.HTTPConnectionManager,
		ConfigType: &listener.Filter_TypedConfig{
			TypedConfig: managerConfig,
		},
	}
}

func (b *FilterChainBuilder) buildDownstreamTransportSocket() *core.TransportSocket {
	c := &auth.DownstreamTlsContext{
		CommonTlsContext: &auth.CommonTlsContext{
			AlpnProtocols:                  []string{"h2", "http/1.1"},
			TlsCertificateSdsSecretConfigs: []*auth.SdsSecretConfig{b.sdsCertificateConfig},
			TlsParams: &auth.TlsParameters{
				TlsMinimumProtocolVersion: auth.TlsParameters_TLSv1_2,
			},
		},
	}
	tlsc, _ := anypb.New(c)

	return &core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &core.TransportSocket_TypedConfig{
			TypedConfig: tlsc,
		},
	}
}
