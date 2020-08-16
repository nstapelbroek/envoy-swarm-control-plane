package provider

import (
	"context"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

type ADS interface {
	Provide(ctx context.Context) (clusters, listeners []types.Resource, err error)
}

type SDS interface {
	Provide(ctx context.Context) (secrets []types.Resource, err error)
	GetCertificateConfigKey(vhost *route.VirtualHost) string
}
