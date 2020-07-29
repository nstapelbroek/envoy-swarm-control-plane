package provider

import (
	"context"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

type Provider interface {
	ProvideClustersAndListeners(ctx context.Context) (clusters, listeners []types.Resource, err error)
}
