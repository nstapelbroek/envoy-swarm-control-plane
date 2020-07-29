package producer

import "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/discovery"

func InitialStartup(dispatchChannel chan discovery.Reason) {
	dispatchChannel <- "Initial startup"
}
