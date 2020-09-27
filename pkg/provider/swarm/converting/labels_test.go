package converting

import (
	"testing"

	types "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"gotest.tools/assert"
)

func TestServiceLabelDefaults(t *testing.T) {
	defaults := NewServiceLabel()

	assert.Equal(t, defaults.Route.PathPrefix, "/")
	assert.Check(t, len(defaults.Route.ExtraDomains) == 0)
	assert.Equal(t, defaults.Route.Domain, "")

	assert.Equal(t, defaults.Endpoint.Protocol, types.SocketAddress_TCP)
	assert.Equal(t, defaults.Endpoint.Port, types.SocketAddress_PortValue{PortValue: 0})
}

func TestServiceLabelInvalidDNS(t *testing.T) {
	label := NewServiceLabel()
	label.Endpoint.Port = types.SocketAddress_PortValue{PortValue: 80}
	label.Route.Domain = "*"

	assert.Error(t, label.Validate(), "the route.domain is not a valid DNS name")
}
