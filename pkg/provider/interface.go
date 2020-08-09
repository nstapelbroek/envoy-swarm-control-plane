package provider

import (
	"context"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

type Resource interface {
	Provide(ctx context.Context) (clusters, listeners []types.Resource, err error)
}

type TLS interface {
	HasCertificate(vhost route.VirtualHost) bool
	GetCertificate(vhost route.VirtualHost) interface{}
	IssueCertificate(vhost route.VirtualHost) interface{}
}
