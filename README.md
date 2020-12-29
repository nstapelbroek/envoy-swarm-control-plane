# Envoy Swarm Control Plane
Opinionated control plane software that enables virtual hosting in docker swarm by using Envoy as an edge proxy.

[![asciicast](https://asciinema.org/a/LEu3l3sLfIVVA6GomAh5cn0Mo.svg)](https://asciinema.org/a/LEu3l3sLfIVVA6GomAh5cn0Mo)

## Features

- Made for Docker Swarm 
  - Discovers service configuration without any additional software
  - Relies on swarms routing mesh to proxy traffic to services
  - Reads configuration from deployment labels
  - Instantly detects changes in stack configurations
  - Gives you freedom to run your edge proxies on worker nodes
- SSL/TLS support
  - Redirect HTTP to HTTPS
  - TLS enabled vhosts will offer HTTP/1.1 and HTTP/2
  - TLS 1.2 and up
- LetsEncrypt integration
  - For one or multiple (bundled) domains
  - Automatic renewals
- Able to store certificates on Disk or Object storage
- Tries to play nice with system resources
  - So far it uses ~25mb on a swarm with 20 swarm services

## Getting started
Use the [docs](docs/introduction.md) to learn more.
  
## Roadmap:
I'm working to get this to an MVP state.
You can follow the progress in [the project board on Github](https://github.com/nstapelbroek/envoy-swarm-control-plane/projects/1). 