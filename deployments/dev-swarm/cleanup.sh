#!/usr/bin/env bash

docker stack rm envoy || docker network rm edge-traffic
