package swarm

import (
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/swarm/converting"
)

type ListenerBuilder struct {
	sdsProvider     provider.SDS
	acmeIntegration *acme.Integration
}

func NewListenerBuilder(sdsProvider provider.SDS, acmeIntegration *acme.Integration) *ListenerBuilder {
	return &ListenerBuilder{
		sdsProvider:     sdsProvider,
		acmeIntegration: acmeIntegration,
	}
}

// ProvideListeners breaks down a vhost collection into listener configs. It will enable TLS + redirect when the
// matching certificates are found or just keep communication HTTP and optionally start an ACME HTTP-01 challenge
func (l *ListenerBuilder) ProvideListeners(collection *converting.VhostCollection) (listeners []types.Resource, err error) {
	httpBuilder := converting.NewListenerBuilder("http_listener")
	httpsBuilder := converting.NewListenerBuilder("https_listener").EnableTLS()
	httpFilter := converting.NewFilterChainBuilder("httpFilter")

	for i := range collection.Vhosts {
		vhost := collection.Vhosts[i]

		if l.sdsProvider.HasCertificate(vhost) {
			httpFilter.ForVhost(converting.CreateRedirectVhost(vhost))
			httpsBuilder.AddFilterChain(l.createHTTPSFilterChain(vhost))
			continue
		}

		if l.acmeIntegration != nil {
			vhost.Routes = append(vhost.Routes, l.acmeIntegration.GetHTTP01Route())
		}

		httpFilter.ForVhost(vhost)
	}

	httpListener := httpBuilder.AddFilterChain(httpFilter).Build()
	if err := httpListener.Validate(); err != nil {
		return nil, err
	}

	listeners = append(listeners, httpBuilder.AddFilterChain(httpFilter).Build())
	httpsListener := httpsBuilder.Build()
	if len(httpsListener.FilterChains) == 0 {
		// Returning a listener without filter chains does not work, guess we are only serving HTTP
		return listeners, nil
	}

	listeners = append(listeners, httpsListener)
	if err := httpsListener.Validate(); err != nil {
		return nil, err
	}

	return listeners, nil
}

func (l *ListenerBuilder) createHTTPSFilterChain(vhost *route.VirtualHost) *converting.FilterChainBuilder {
	return converting.NewFilterChainBuilder(vhost.Name).EnableTLS(vhost.Domains, l.sdsProvider.GetCertificateConfig(vhost)).ForVhost(vhost)
}
