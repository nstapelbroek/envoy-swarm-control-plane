# Envoy Swarm Control Plane
Opinionated control plane software that configures [Envoy Proxy](https://www.envoyproxy.io/) as a reverse proxy for 
docker swarm. Enable vhosting accross your swarm with just two services!

[![asciicast](https://asciinema.org/a/LEu3l3sLfIVVA6GomAh5cn0Mo.svg)](https://asciinema.org/a/LEu3l3sLfIVVA6GomAh5cn0Mo)

## Features

- Made for Docker Swarm 
  - Discovers service configuration without any additional software
  - Relies on swarms routing mesh to proxy traffic to services
  - Reads configuration from deployment labels
  - Instantly detects changes in stack configurations
  - Gives you freedom to run your reverse proxies on worker nodes
- SSL/TLS support
  - Redirect HTTP to HTTPS
  - TLS enabled vhosts will offer HTTP/1.1 and HTTP/2
  - TLS 1.2 and up
- LetsEncrypt integration
  - For one or multiple (bundled) domains
  - Automatic renewals
- Able to store certificates on Disk or S3/Object storage
- Tries to play nice with system resources
  - So far it uses ~25mb on a swarm cluster with 20 services

## Getting started
Use the [docs](https://nstapelbroek.github.io/envoy-swarm-control-plane/) to learn more.
