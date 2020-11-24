#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

docker build -t nstapelbroek/envoy-edge:0.15 "$DIR"/envoy/docker
docker buildx build -t nstapelbroek/envoy-swarm-control-plane:0.15 "$DIR/../../"