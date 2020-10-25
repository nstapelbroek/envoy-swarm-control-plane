#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

docker network create --driver=overlay --attachable edge-traffic
docker stack deploy --compose-file "${DIR}"/stack.yml envoy