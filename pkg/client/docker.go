package client

import docker "github.com/docker/docker/client"

func NewDockerClient() *docker.Client {
	httpHeaders := map[string]string{
		"User-Agent": "Envoy Swarm Control Plane",
	}

	c, err := docker.NewClientWithOpts(
		docker.FromEnv,
		docker.WithHTTPHeaders(httpHeaders),
	)
	if err != nil {
		panic(err)
	}

	return c
}
