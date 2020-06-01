package main

import (
	"context"
	"flag"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal"
	"log"
)

var (
	debug  bool
	port   uint
	nodeID string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Use debug logging")
	flag.UintVar(&port, "port", 9876, "Management server port")
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")
}

func main() {
	flag.Parse()
	ctx := context.Background()

	config := cache.NewSnapshotCache(true, cache.IDHash{}, logger{})
	s := server.NewServer(context.Background(), config, nil)

	// start the xDS server
	go internal.RunGRPCServer(ctx, s, port)

	var clusters, endpoints, routes, listeners, runtimes []types.Resource
	snapshot := cache.NewSnapshot("1.0", endpoints, clusters, routes, listeners, runtimes)

	_ = config.SetSnapshot(nodeID, snapshot)

	select {}
}

type logger struct{}

func (logger logger) Debugf(format string, args ...interface{}) {
	if debug {
		log.Printf(format+"\n", args...)
	}
}

func (logger logger) Infof(format string, args ...interface{}) {
	if debug {
		log.Printf(format+"\n", args...)
	}
}

func (logger logger) Warnf(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}

func (logger logger) Errorf(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}
