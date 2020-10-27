#!/usr/bin/env bash

docker network create --driver=overlay --attachable edge-traffic
docker stack deploy --compose-file "${DIR}"envoy/stack.yml envoy
docker stack deploy --compose-file "${DIR}"userland-service/stack.yml example