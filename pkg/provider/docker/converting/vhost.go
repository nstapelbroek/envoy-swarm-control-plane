package converting

import (
	"fmt"

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
	virtualHost.Routes = append(virtualHost.Routes, w.createRoute(clusterIdentifier, labels))
	for i := range extraDomains {
		virtualHost.Domains = append(virtualHost.Domains, extraDomains[i])
		w.usedDomains[extraDomains[i]] = virtualHost
	}

	return nil
}

func (w VhostCollection) createRoute(clusterIdentifier string, labels *ServiceLabel) *route.Route {
	return &route.Route{
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
}

func CreateRedirectVhost(originalVhost *route.VirtualHost) *route.VirtualHost {
	return &route.VirtualHost{
		Name:    originalVhost.Name,
		Domains: originalVhost.Domains,
		Routes: []*route.Route{{
			Name: "https_redirect",
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: "/",
				},
			},
			Action: &route.Route_Redirect{
				Redirect: &route.RedirectAction{
					SchemeRewriteSpecifier: &route.RedirectAction_HttpsRedirect{
						HttpsRedirect: true,
					},
				},
			},
		}},
	}
}
