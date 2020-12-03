package converting

import (
	"testing"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/stretchr/testify/assert"
)

func assertAddress(t *testing.T, expected string, listener *listener.Listener) {
	assert.IsType(t, &core.Address_SocketAddress{}, listener.Address.Address)
	address := listener.Address.Address.(*core.Address_SocketAddress)

	assert.Equal(t, expected, address.SocketAddress.Address)
}

func assertPort(t *testing.T, expected uint32, listener *listener.Listener) {
	assert.IsType(t, &core.Address_SocketAddress{}, listener.Address.Address)
	address := listener.Address.Address.(*core.Address_SocketAddress)
	assert.IsType(t, &core.SocketAddress_PortValue{}, address.SocketAddress.PortSpecifier)
	port := address.SocketAddress.PortSpecifier.(*core.SocketAddress_PortValue)
	assert.Equal(t, expected, port.PortValue)
}

func TestDefaultListenerIsForHttp(t *testing.T) {
	subject := NewListenerBuilder("some_builder")

	result := subject.Build()

	assertAddress(t, "0.0.0.0", result)
	assertPort(t, uint32(80), result)
	assert.Len(t, result.FilterChains, 0)
	assert.Len(t, result.ListenerFilters, 0)
}

func TestEnableTLS(t *testing.T) {
	subject := NewListenerBuilder("some_builder")

	result := subject.EnableTLS().Build()

	assertAddress(t, "0.0.0.0", result)
	assertPort(t, uint32(443), result)
	assert.Len(t, result.FilterChains, 0)
	assert.Len(t, result.ListenerFilters, 1)
	assert.Equal(t, "envoy.filters.listener.tls_inspector", result.ListenerFilters[0].Name)
}

func TestAddFilterChain(t *testing.T) {
	subject := NewListenerBuilder("some_builder")

	result := subject.AddFilterChain(NewFilterChainBuilder("some_tls_matcher_chain")).Build()

	assert.Len(t, result.FilterChains, 1)
}

func TestDownstreamListenerBufferSize(t *testing.T) {
	subject := NewListenerBuilder("some_builder")

	result := subject.Build()

	assert.Equal(t, uint32(32768), result.PerConnectionBufferLimitBytes.Value)
}
