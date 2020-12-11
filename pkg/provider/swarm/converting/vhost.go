package converting

import (
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"time"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type VhostCollection struct {
	Vhosts      map[string]*route.VirtualHost
	usedDomains map[string]*route.VirtualHost
}

func NewVhostCollection() *VhostCollection {
	return &VhostCollection{
		Vhosts:      make(map[string]*route.VirtualHost),
		usedDomains: make(map[string]*route.VirtualHost),
	}
}

func (w VhostCollection) AddService(clusterIdentifier string, labels *ServiceLabel) (err error) {
	primaryDomain := labels.Route.Domain

	virtualHost, exist := w.Vhosts[primaryDomain]
	if !exist {
		if _, exist := w.usedDomains[primaryDomain]; exist {
			return fmt.Errorf("domain %s is already used in another vhost", primaryDomain)
		}

		virtualHost = &route.VirtualHost{
			Name:    primaryDomain,
			Domains: []string{primaryDomain},
			Routes:  []*route.Route{},
		}
	}

	// Calculate and validate changes to the vhost domains
	var extraDomains []string
	for _, extraDomain := range labels.Route.ExtraDomains {
		if extraDomain == primaryDomain {
			continue
		}

		if v, exist := w.usedDomains[extraDomain]; exist {
			if v != virtualHost {
				return fmt.Errorf("domain %s is already used in another vhost", extraDomain)
			}
			continue
		}
		extraDomains = append(extraDomains, extraDomain)
	}

	// Validation ended above, applying changes
	w.Vhosts[primaryDomain] = virtualHost
	w.usedDomains[primaryDomain] = virtualHost

	// Order of routes matter, ensure that the default catch-all / comes last
	newRoute := w.createRoute(clusterIdentifier, labels)
	if labels.Route.PathPrefix == "/" {
		virtualHost.Routes = append(virtualHost.Routes, newRoute)
	} else {
		virtualHost.Routes = append([]*route.Route{newRoute}, virtualHost.Routes...)
	}

	for i := range extraDomains {
		virtualHost.Domains = append(virtualHost.Domains, extraDomains[i])
		w.usedDomains[extraDomains[i]] = virtualHost
	}

	return nil
}

func (w VhostCollection) createRoute(clusterIdentifier string, labels *ServiceLabel) *route.Route {
	const clientIdleTimeout = 15

	return &route.Route{
		Name: clusterIdentifier + "_route",
		Match: &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: labels.Route.PathPrefix,
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: clusterIdentifier,
				},
				// https://github.com/envoyproxy/envoy/issues/8517#issuecomment-540225144
				IdleTimeout: ptypes.DurationProto(clientIdleTimeout * time.Second),
				Timeout:     ptypes.DurationProto(labels.Endpoint.RequestTimeout),
			},
		},
	}
}
