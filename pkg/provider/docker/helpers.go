package docker

import (
	swarmtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/docker/converting"
)

func inIngressNetwork(service *swarm.Service, ingress *swarmtypes.NetworkResource) bool {
	for _, vip := range service.Endpoint.VirtualIPs {
		if vip.NetworkID == ingress.ID {
			return true
		}
	}

	return false
}

func mapVhostsToHTTPListener(collection *converting.VhostCollection) types.Resource {
	filter := converting.NewFilterChainBuilder("http")
	for i := range collection.Vhosts {
		filter.ForVhost(collection.Vhosts[i])
	}

	return converting.NewListenerBuilder("http_listener").AddFilterChain(filter).Build()
}

func mapVhostsToHttpsListeners(collection *converting.VhostCollection, sdsProvider provider.SDS) []types.Resource {
	httpListener := converting.NewListenerBuilder("http_listener")
	httpsListener := converting.NewListenerBuilder("https_listener").EnableTLS()
	httpFilter := converting.NewFilterChainBuilder("httpFilter")

	for i := range collection.Vhosts {
		vhost := collection.Vhosts[i]

		// if there is no certificate, serve using http
		if !sdsProvider.HasCertificate(vhost) {
			httpFilter.ForVhost(vhost)
			continue
		}

		httpFilter.ForVhost(converting.CreateRedirectVhost(vhost))
		httpsListener.AddFilterChain(
			converting.NewFilterChainBuilder(vhost.Name).
				EnableTLS(vhost.Domains, sdsProvider.GetCertificateConfig(vhost)).
				ForVhost(vhost),
		)
	}

	return []types.Resource{
		httpListener.AddFilterChain(httpFilter).Build(),
		httpsListener.Build(),
	}
}
