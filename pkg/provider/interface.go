// A provider is the bridge between xDS objects and external services or certificates

package provider

import (
	"context"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

type ADS interface {
	Provide(ctx context.Context) (clusters, listeners []types.Resource, err error)
}

type SDS interface {
	Provide(ctx context.Context) (secrets []types.Resource, err error)
	HasValidCertificate(vhost *route.VirtualHost) bool
	GetCertificateConfig(vhost *route.VirtualHost) *auth.SdsSecretConfig
}
