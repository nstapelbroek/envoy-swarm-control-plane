#! /bin/sh
docker stack rm envoy myapi myfrontend
docker network rm edge-traffic
sleep 10

docker network create --driver=overlay --attachable edge-traffic
docker stack deploy --compose-file ./envoy-stack.yml envoy
docker stack deploy --compose-file ./api-stack.yml myapi
docker stack deploy --compose-file ./frontend-stack.yml myfrontend
