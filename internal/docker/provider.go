package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

type SwarmProvider struct {
	dockerClient client.APIClient
}

func NewSwarmProvider() SwarmProvider {
	httpHeaders := map[string]string{
		"User-Agent": "Envoy Swarm Control Plane",
	}

	c, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithHTTPHeaders(httpHeaders),
	)
	if err != nil {
		panic(err)
	}

	return SwarmProvider{dockerClient: c}
}

// ListServices will convert swarm service definitions to our own Models
func (s SwarmProvider) ListServices(ctx context.Context) ([]swarm.Service, error) {
	return s.dockerClient.ServiceList(ctx, types.ServiceListOptions{})
}
