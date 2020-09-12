package watcher

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/client"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/snapshot"
)

type SwarmEvent struct {
	client docker.APIClient
	logger logger.Logger
}

func ForSwarmEvent(log logger.Logger) *SwarmEvent {
	return &SwarmEvent{client: client.NewDockerClient(), logger: log}
}

func (s SwarmEvent) Watch(ctx context.Context, dispatchChannel chan snapshot.UpdateReason) {
	events, errorEvent := s.client.Events(ctx, types.EventsOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "type", Value: "service"}),
	})

	for {
		select {
		case <-events:
			s.logger.Debugf("received service event from docker")
			dispatchChannel <- "swarm event" // todo investigate if waiting for a receiver here blocks other interactions.
		case err := <-errorEvent:
			s.logger.Errorf(err.Error())
			s.Watch(ctx, dispatchChannel) // Auto recover on errors @see github.com/docker/engine/docker/events.go:19
		}
	}
}
