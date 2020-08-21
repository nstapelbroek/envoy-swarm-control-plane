#! /bin/sh
docker stack rm envoy myapi myfrontend
docker network rm edge-traffic
