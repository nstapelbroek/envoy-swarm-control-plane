package main

import (
	"context"
	"flag"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/internal"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	debug  bool
	port   uint
	pollInterval   uint
	nodeID string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Use debug logging")
	flag.UintVar(&port, "port", 9876, "Management server port")
	flag.UintVar(&pollInterval, "interval", 15, "Poll interval")
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")
}

func main() {
	flag.Parse()
	ctx := internal.WithLogger(context.Background())

	c := cache.NewSnapshotCache(true, cache.IDHash{}, logger{})
	// Pass in new context here since i rather stop the gRPC server and don't bother with this
	s := server.NewServer(context.Background(), c, nil)
	p := internal.NewSwarmProvider()



	go internal.RunGRPCServer(ctx, s, port)
	go internal.StartPollingForChanges(ctx, p, c)


	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
		ctx.Done()
		return
	}
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
