# Envoy Swarm Control Plane
Opinionated control plane software that enables virtual hosting in docker swarm by using Envoy as an edge proxy.

todo: logo or small demo

## Features

- Made for Docker Swarm 
  - Discovers service configuration without any additional software
  - Relies on swarms routing mesh (DNSRR or VIP) to proxy traffic to services
  - Reads configuration from service labels
  - Instantly detects changes in stack configurations
- Automatic SSL/TLS certificates for each virtual host
  - Supports bundeled domain names for a single certificate
  - Redirect from HTTP to HTTPS
- Enables you to route traffic via swarm worker nodes instead of a manager
- Tries to play nice with system resources

## Getting started
Use the [docs](docs/introduction.md) to learn more.
  
## Roadmap:
I'm working to get this to an MVP state. You can follow the progress in [the project board on Github](https://github.com/nstapelbroek/envoy-swarm-control-plane/projects/1). 