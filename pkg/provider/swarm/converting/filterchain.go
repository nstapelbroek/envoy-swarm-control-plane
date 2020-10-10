package converting

import (
	"fmt"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
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
	routeType := "http"
	if b.configureTLS {
		routeType = "https"
	}

	routes := &route.RouteConfiguration{
		Name:         fmt.Sprintf("%s_%s_routes", b.name, routeType),
		VirtualHosts: b.vhosts,
	}
	conManager := &hcm.HttpConnectionManager{
		CodecType:      hcm.HttpConnectionManager_AUTO,
		StatPrefix:     b.name,
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{RouteConfig: routes},
		HttpFilters:    []*hcm.HttpFilter{{Name: "envoy.filters.http.router"}},
	}
	managerConfig, err := ptypes.MarshalAny(conManager)
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
		},
	}
	tlsc, _ := ptypes.MarshalAny(c)

	return &core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &core.TransportSocket_TypedConfig{
			TypedConfig: tlsc,
		},
	}
}
