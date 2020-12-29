package swarm

import (
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/swarm/converting"
)

type ListenerProvider struct {
	sdsProvider     provider.SDS
	acmeIntegration *acme.Integration
}

func NewListenerProvider(sdsProvider provider.SDS, acmeIntegration *acme.Integration) *ListenerProvider {
	return &ListenerProvider{
		sdsProvider:     sdsProvider,
		acmeIntegration: acmeIntegration,
	}
}

// ProvideListeners breaks down a vhost collection into listener configs it will return a collection of max 2 listeners
// for port 80 and 443.
func (l *ListenerProvider) ProvideListeners(collection *converting.VhostCollection) ([]types.Resource, error) {
	httpListener, httpsListener := l.createListenersFromVhosts(collection)
	if err := httpListener.Validate(); err != nil {
		return nil, err
	}

	if len(httpsListener.FilterChains) == 0 || httpsListener.Validate() != nil {
		return []types.Resource{httpListener}, nil
	}

	return []types.Resource{httpListener, httpsListener}, nil
}

// createListenersFromVhosts will create so-called network filter chains for each vhost that has a TLS certificate.
// This assures we serve the correct certificate even before we know what the request speaks (we assume HTTP)
func (l *ListenerProvider) createListenersFromVhosts(collection *converting.VhostCollection) (http, https *listener.Listener) {
	// Every vhost that doesn't have a certificate will end up in our generic HTTP catch-all filter
	httpFilter := converting.NewFilterChainBuilder("httpFilter")

	// Each filter is added to a listener, a listener will instruct envoy to "listen" on a port
	httpBuilder := converting.NewListenerBuilder("http_listener")
	httpsBuilder := converting.NewListenerBuilder("https_listener").EnableTLS()

	for i := range collection.Vhosts {
		vhost := collection.Vhosts[i]
		hasValidCertificate := false
		if l.sdsProvider != nil {
			hasValidCertificate = l.sdsProvider.HasValidCertificate(vhost)
		}

		// handle LetsEncrypt first because it might mutate the vhost config
		if l.acmeIntegration != nil {
			// this covers two use cases: new certificates (hasValidCertificate) and schedules renewals (IsScheduledForIssuing)
			if !hasValidCertificate || l.acmeIntegration.IsScheduledForIssuing(vhost) {
				vhost = l.acmeIntegration.PrepareVhostForIssuing(vhost)
			}

			if hasValidCertificate {
				l.acmeIntegration.EnableAutoRenewal(vhost)
			}
		}

		if hasValidCertificate {
			httpsFilter := l.createFilterChainWithTLS(vhost)
			httpsFilter.ForVhost(vhost)
			httpsBuilder.AddFilterChain(httpsFilter)

			// note that our redirect logic plays nice with paths, so ACME challenges should still work
			httpFilter.ForVhost(createNewHTTPSRedirectVhost(vhost))
		} else {
			httpFilter.ForVhost(vhost)
		}
	}

	httpBuilder.AddFilterChain(httpFilter)
	return httpBuilder.Build(), httpsBuilder.Build()
}

func (l *ListenerProvider) createFilterChainWithTLS(vhost *route.VirtualHost) *converting.FilterChainBuilder {
	return converting.NewFilterChainBuilder(vhost.Name).EnableTLS(vhost.Domains, l.sdsProvider.GetCertificateConfig(vhost))
}

func createNewHTTPSRedirectVhost(originalVhost *route.VirtualHost) *route.VirtualHost {
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
