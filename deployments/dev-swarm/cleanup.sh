#! /bin/sh
docker stack rm envoy || docker network rm edge-traffic
