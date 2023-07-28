package converting

import (
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tls_inspectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ListenerBuilder struct {
	name         string
	configureTLS bool
	filterChains []*FilterChainBuilder
}

func NewListenerBuilder(name string) *ListenerBuilder {
	return &ListenerBuilder{
		name:         name,
		configureTLS: false,
		filterChains: []*FilterChainBuilder{},
	}
}

func (b *ListenerBuilder) EnableTLS() *ListenerBuilder {
	b.configureTLS = true

	return b
}

func (b *ListenerBuilder) AddFilterChain(f *FilterChainBuilder) *ListenerBuilder {
	// We could validate here if they are both aligned for tls configuration, but meh. it's a programmer error
	b.filterChains = append(b.filterChains, f)

	return b
}

func (b *ListenerBuilder) Build() *listener.Listener {
	const PerConnectionBufferLimit = 32768 // 32 KiB

	port := 80
	var listenerFilters []*listener.ListenerFilter

	if b.configureTLS {
		port = 443
		tlsFilterConfig, _ := anypb.New(&tls_inspectorv3.TlsInspector{})
		listenerFilters = []*listener.ListenerFilter{{
			Name:       wellknown.TLSInspector,
			ConfigType: &listener.ListenerFilter_TypedConfig{TypedConfig: tlsFilterConfig},
		}}
	}

	var chains []*listener.FilterChain
	for i := range b.filterChains {
		chains = append(chains, b.filterChains[i].Build())
	}

	return &listener.Listener{
		Name: b.name,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: uint32(port),
					},
				},
			},
		},
		ListenerFilters:               listenerFilters,
		FilterChains:                  chains,
		PerConnectionBufferLimitBytes: &wrapperspb.UInt32Value{Value: uint32(PerConnectionBufferLimit)},
	}
}
