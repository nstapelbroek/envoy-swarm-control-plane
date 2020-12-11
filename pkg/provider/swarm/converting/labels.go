package converting

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	valid "github.com/asaskevich/govalidator"
	types "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

type ServiceEndpoint struct {
	RequestTimeout time.Duration
	Protocol       types.SocketAddress_Protocol
	Port           types.SocketAddress_PortValue
}

type ServiceRoute struct {
	Domain       string
	ExtraDomains []string // Note that you cannot assume that the domain or extraDomains are valid and reachable
	PathPrefix   string
}

type ServiceLabel struct {
	Endpoint ServiceEndpoint
	Route    ServiceRoute
}

func (l *ServiceLabel) setEndpointProp(property, value string) {
	switch strings.ToLower(property) {
	case "timeout":
		timeout, _ := time.ParseDuration(value)
		l.Endpoint.RequestTimeout = timeout
	case "protocol":
		p := types.SocketAddress_TCP
		if strings.EqualFold(value, "udp") {
			p = types.SocketAddress_UDP
		}

		l.Endpoint.Protocol = p
	case "port":
		v, _ := strconv.ParseUint(value, 10, 32)
		l.Endpoint.Port = types.SocketAddress_PortValue{
			PortValue: uint32(v),
		}
	}
}

func (l *ServiceLabel) setRouteProp(property, value string) {
	switch strings.ToLower(property) {
	case "path":
		l.Route.PathPrefix = fmt.Sprintf("/%s", strings.TrimPrefix(value, "/"))
	case "domain":
		l.Route.Domain = value
	case "extra-domains":
		l.Route.ExtraDomains = strings.Split(value, ",")
	}
}

var serviceLabelRegex = regexp.MustCompile(`(?Uim)envoy\.(?P<type>\S+)\.(?P<property>\S+$)`)

// NewServiceLabel will create an ServiceLabel with default values
// Note that some values like the timeouts are based upon best practices for running at the edge: https://www.envoyproxy.io/docs/envoy/latest/configuration/best_practices/edge
func NewServiceLabel() ServiceLabel {
	return ServiceLabel{
		ServiceEndpoint{
			RequestTimeout: 15 * time.Second,
			Protocol:       types.SocketAddress_TCP,
			Port:           types.SocketAddress_PortValue{PortValue: 0},
		},
		ServiceRoute{
			ExtraDomains: []string{},
			PathPrefix:   "/",
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

func (l ServiceLabel) Validate() error {
	if l.Endpoint.Port.PortValue <= 0 {
		return errors.New("there is no endpoint.port label specified")
	}

	if l.Route.Domain == "" {
		return errors.New("there is no route.domain label specified")
	}

	if l.Endpoint.RequestTimeout.Seconds() < 0 {
		return errors.New("the endpoint.timeout can't be a negative number")
	}

	if !valid.IsDNSName(l.Route.Domain) {
		return errors.New("the route.domain is not a valid DNS name")
	}

	for i := range l.Route.ExtraDomains {
		if !valid.IsDNSName(l.Route.ExtraDomains[i]) {
			return errors.New("the route.extra-domains contains an invalid DNS name")
		}
	}

	return nil
}
