package conversion

import (
	"errors"
	"fmt"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
)

type WebListener struct {
	vhosts map[string]*route.VirtualHost
}

func NewWebListener() *WebListener {
	return &WebListener{
		vhosts: make(map[string]*route.VirtualHost),
	}
}

func (w WebListener) AddRoute(clusterIdentifier string, labels *ServiceLabel) (err error) {
	primaryDomain := labels.Route.Domain
	for _, extraDomain := range labels.Route.ExtraDomains {
		if err := w.validateExtraDomain(extraDomain); err != nil {
			return err
		}
	}

	virtualHost, exist := w.vhosts[primaryDomain]
	if !exist {
		virtualHost = &route.VirtualHost{
			Name:    primaryDomain,
			Domains: []string{primaryDomain},
			Routes:  []*route.Route{},
		}
	}

	virtualHost.Domains = append(virtualHost.Domains, labels.Route.ExtraDomains...)

	newRoute := &route.Route{
		Name: clusterIdentifier + "_route",
		Match: &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: labels.Route.Path,
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: clusterIdentifier,
				},
			},
		},
	}
	if err := newRoute.Validate(); err != nil {
		return err
	}

	virtualHost.Routes = append(virtualHost.Routes, newRoute)
	w.vhosts[primaryDomain] = virtualHost

	return nil
}

func (w WebListener) BuildListener() *listener.Listener {
	mngr, err := ptypes.MarshalAny(w.buildConnectionManger())
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name: "http_listener",
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0", // Default to all addresses since we don't know where our proxies are running
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: 80, // todo
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: mngr,
				},
			}},
		}},
	}
}

func (w WebListener) validateExtraDomain(domain string) error {
	if domain == "*" {
		return errors.New("wildcard cannot be used in an extra domain")
	}

	if w.vhosts[domain] != nil {
		return fmt.Errorf("domain %s is already used as a primary domain in another vhost", domain)
	}

	return nil
}

func (w WebListener) buildRouteConfig() *route.RouteConfiguration {
	r := &route.RouteConfiguration{Name: "swarm_routes"}
	for _, host := range w.vhosts {
		r.VirtualHosts = append(r.VirtualHosts, host)
	}

	return r
}

func (w WebListener) buildConnectionManger() *hcm.HttpConnectionManager {
	return &hcm.HttpConnectionManager{
		CodecType:      hcm.HttpConnectionManager_AUTO,
		StatPrefix:     "http",
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{RouteConfig: w.buildRouteConfig()},
		HttpFilters:    []*hcm.HttpFilter{{Name: "envoy.filters.http.router"}},
	}
}
