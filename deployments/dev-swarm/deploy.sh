#! /bin/sh
docker network inspect edge-traffic || docker network create --driver=overlay edge-traffic

docker stack rm envoy
docker stack deploy --compose-file ./envoy-stack.yml envoy

docker stack rm myapi
docker stack deploy --compose-file ./api-stack.yml myapi

docker stack rm myfrontend
docker stack deploy --compose-file ./frontend-stack.yml myfrontend
