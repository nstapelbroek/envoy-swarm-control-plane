package producer

import (
	"context"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/discovery"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/docker"
)

type SwarmEventProducer struct {
	provider docker.SwarmProvider
	logger   logger.Logger
}

func NewSwarmEventProducer(provider docker.SwarmProvider, log logger.Logger) *SwarmEventProducer {
	return &SwarmEventProducer{provider: provider, logger: log}
}

func (s SwarmEventProducer) UpdateOnSwarmEvents(ctx context.Context, dispatchChannel chan discovery.Reason) {
	// Drilling down the context as calling done will close the channels for us
	events, errorEvent := s.provider.Events(ctx)

	for {
		select {
		case <-events:
			s.logger.Debugf("received service event from docker")
			dispatchChannel <- "swarm event" // todo investigate if waiting for a receiver here blocks other interactions with the docker client.
		case err := <-errorEvent:
			s.logger.Errorf(err.Error())
			s.UpdateOnSwarmEvents(ctx, dispatchChannel) // Auto recover on errors @see github.com/docker/engine/client/events.go:19
		}
	}
}
