#! /bin/sh
docker network create --driver=overlay --attachable edge-traffic
docker stack deploy --compose-file ./stack.yml envoy