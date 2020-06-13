package docker

import (
"context"
"github.com/docker/docker/api/types"
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

func (s SwarmProvider) Services(ctx context.Context) (interface{}, error) {
	serviceList, err := s.dockerClient.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return nil, err
	}

	print(serviceList)

	return serviceList, nil
}