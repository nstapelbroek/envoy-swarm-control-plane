package converting

import (
	"testing"
	"time"

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

func TestServiceLabelAutoPathPrefix(t *testing.T) {
	label := NewServiceLabel()
	label.setRouteProp("path", "api")

	assert.Equal(t, label.Route.PathPrefix, "/api")
}

func TestServiceLabelInvalidDNS(t *testing.T) {
	label := NewServiceLabel()
	label.Endpoint.Port = types.SocketAddress_PortValue{PortValue: 80}
	label.Route.Domain = "*"

	assert.Error(t, label.Validate(), "the route.domain is not a valid DNS name")
}

func TestServiceLabelInvalidTimeout(t *testing.T) {
	label := NewServiceLabel()
	label.Route.Domain = "example"
	label.Endpoint.Port = types.SocketAddress_PortValue{PortValue: 80}
	label.Endpoint.RequestTimeout = -500 * time.Millisecond

	assert.Error(t, label.Validate(), "the endpoint.timeout can't be a negative number")
}

func TestParseServiceLabelsEndpointTimeout(t *testing.T) {
	labels := make(map[string]string)
	labels["envoy.endpoint.timeout"] = "30m"

	parsed := ParseServiceLabels(labels)

	assert.Equal(t, parsed.Endpoint.RequestTimeout.Seconds(), float64(1800))
}
