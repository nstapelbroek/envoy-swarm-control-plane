#! /bin/sh
docker network inspect edge-traffic || docker network create --driver=overlay edge-traffic

docker stack rm envoy
docker stack deploy --compose-file ./envoy-stack.yml envoy

docker stack rm myapp
docker stack deploy --compose-file ./app-stack.yml myapp
