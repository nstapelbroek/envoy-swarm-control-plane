package converting

import (
	"errors"
	"fmt"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type VhostCollection struct {
	Vhosts map[string]*route.VirtualHost
}

func NewVhostCollection() *VhostCollection {
	return &VhostCollection{
		Vhosts: make(map[string]*route.VirtualHost),
	}
}

func (w VhostCollection) AddRoute(clusterIdentifier string, labels *ServiceLabel) (err error) {
	primaryDomain := labels.Route.Domain
	for _, extraDomain := range labels.Route.ExtraDomains {
		if err := w.validateExtraDomain(extraDomain); err != nil {
			return err
		}
	}

	virtualHost, exist := w.Vhosts[primaryDomain]
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
	w.Vhosts[primaryDomain] = virtualHost

	return nil
}

func (w VhostCollection) validateExtraDomain(domain string) error {
	if domain == "*" {
		return errors.New("wildcard cannot be used in an extra domain")
	}

	if w.Vhosts[domain] != nil {
		return fmt.Errorf("domain %s is already used as a primary domain in another vhost", domain)
	}

	return nil
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
