package docker

import (
	types "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"regexp"
	"strconv"
	"strings"
)

type ServiceEndpoint struct {
	name     string
	protocol types.SocketAddress_Protocol
	port     types.SocketAddress_PortValue
}

type ServiceLabels struct {
	endpoints map[string]ServiceEndpoint
	// todo add serviceRoutes object here so we know what to use for SNI/Host headers, should maybe be required :thinking:
}

var serviceLabelRegex = regexp.MustCompile(`(?Uim)envoy\.endpoint\.(?P<name>\S+)\.(?P<property>protocol|port)`)

func ParseServiceLabels(labels map[string]string) (s ServiceLabels) {
	s.endpoints = make(map[string]ServiceEndpoint)
	for key, value := range labels {
		if !serviceLabelRegex.MatchString(key) {
			continue
		}

		matches := serviceLabelRegex.FindStringSubmatch(key)
		se, exists := s.endpoints[matches[1]]
		if !exists {
			// Bootstrap default
			se = ServiceEndpoint{
				name:     matches[1],
				protocol: types.SocketAddress_TCP,
			}
		}

		switch matches[2] {
		case "protocol":
			se.protocol = stringToProtocol(value)
		case "port":
			se.port = stringToPort(value)
		}

		s.endpoints[matches[1]] = se
	}

	// strip out any endpoints without a valid port
	for _, endpoint := range s.endpoints {
		if endpoint.port.PortValue == 0 {
			delete(s.endpoints, endpoint.name)
		}
	}

	return s
}

func stringToPort(value string) types.SocketAddress_PortValue {
	v, _ := strconv.ParseUint(value, 10, 32)
	return types.SocketAddress_PortValue{
		PortValue: uint32(v),
	}
}

func stringToProtocol(value string) types.SocketAddress_Protocol {
	if strings.ToLower(value) == "udp" {
		return types.SocketAddress_UDP
	}

	return types.SocketAddress_TCP
}
