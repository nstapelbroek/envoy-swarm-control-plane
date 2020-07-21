package docker

import (
	types "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"regexp"
	"strconv"
	"strings"
)

type ServiceEndpoint struct {
	protocol types.SocketAddress_Protocol
	port     types.SocketAddress_PortValue
}

type ServiceRoute struct {
	domains []string
	path    string
}

type ServiceLabel struct {
	endpoint ServiceEndpoint
	route    ServiceRoute
}

func (l *ServiceLabel) setEndpointProp(property string, value string) {
	switch strings.ToLower(property) {
	case "protocol":
		p := types.SocketAddress_TCP
		if strings.ToLower(value) == "udp" {
			p = types.SocketAddress_UDP
		}

		l.endpoint.protocol = p
	case "port":
		v, _ := strconv.ParseUint(value, 10, 32)
		l.endpoint.port = types.SocketAddress_PortValue{
			PortValue: uint32(v),
		}
	}
}

func (l *ServiceLabel) setRouteProp(property string, value string) {
	switch strings.ToLower(property) {
	case "path":
		l.route.path = value
	case "domains":
		l.route.domains = strings.Split(value, ",")
	}
}

var serviceLabelRegex = regexp.MustCompile(`(?Uim)envoy\.(?P<type>\S+)\.(?P<property>\S+$)`)

// NewServiceLabel will create an ServiceLabel with default values
func NewServiceLabel() ServiceLabel {
	return ServiceLabel{
		ServiceEndpoint{
			protocol: types.SocketAddress_TCP,
			port:     types.SocketAddress_PortValue{PortValue: 0},
		},
		ServiceRoute{
			domains: []string{},
			path:    "/",
		},
	}
}

// ParseServiceLabels constructs a ServiceLabel with default values and passed overrides
func ParseServiceLabels(labels map[string]string) *ServiceLabel {
	s := NewServiceLabel()
	for key, value := range labels {
		if !serviceLabelRegex.MatchString(key) {
			continue
		}

		matches := serviceLabelRegex.FindStringSubmatch(key)
		switch strings.ToLower(matches[1]) { // type
		case "endpoint":
			s.setEndpointProp(matches[2], value)
		case "route":
			s.setRouteProp(matches[2], value)
		}

	}

	return &s
}
