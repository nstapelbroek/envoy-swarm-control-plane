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

func mapVhostsToHttpsListeners(collection *converting.VhostCollection, tlsProvider provider.SDS) []types.Resource {
	httpsListener := converting.NewListenerBuilder("https_listener").EnableTLS()
	redirects := converting.NewFilterChainBuilder("redirects")
	for i := range collection.Vhosts {
		vhost := collection.Vhosts[i]

		configKey := tlsProvider.GetCertificateConfigKey(vhost)
		redirects.ForVhost(converting.CreateRedirectVhost(vhost))
		httpsListener.AddFilterChain(
			converting.NewFilterChainBuilder(vhost.Name).EnableTLS(vhost.Domains, configKey).ForVhost(vhost),
		)
	}

	return []types.Resource{
		converting.NewListenerBuilder("http_listener").AddFilterChain(redirects).Build(),
		httpsListener.Build(),
	}
}
