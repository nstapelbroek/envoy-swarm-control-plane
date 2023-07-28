#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

docker network create --driver=overlay --attachable edge-traffic
docker node update --label-add edge=true "$(docker node ls -q)"
docker stack deploy --compose-file "${DIR}"/envoy/stack.yml envoy
docker stack deploy --compose-file "${DIR}"/http-service/stack.yml example_http