package provider

import (
	"context"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

type Resource interface {
	ProvideClustersAndListener(ctx context.Context) (clusters []types.Resource, listeners types.Resource, err error)
}

type TLS interface {
}
