package watcher

import "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/snapshot"

func CreateInitialStartupEvent(dispatchChannel chan snapshot.UpdateReason) {
	snapshot.RequestUpdate(dispatchChannel, "Initial startup")
}
