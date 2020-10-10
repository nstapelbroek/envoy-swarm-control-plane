# Envoy Swarm Control Plane
Opinionated control plane software that enables virtual hosting in docker swarm by using Envoy as an edge proxy.

todo: logo or small demo

## Features

- Made for Docker Swarm 
  - Discovers service configuration without any additional software
  - Relies on swarms routing mesh (DNSRR or VIP) to proxy traffic to services
  - Reads configuration from deployment labels
  - Instantly detects changes in stack configurations (results in no gateway errors when downscaling)
  - Designed to run the Envoy proxies on worker nodes for more load balancing and availability 
- SSL/TLS support
  - Redirect HTTP to HTTPS automatically
  - TLs enabled vhosts will offer HTTP/1.1 and HTTP/2  
- LetsEncrypt integration
  - For one or multiple (bundled) domains
- Tries to play nice with system resources

## Getting started
Use the [docs](docs/introduction.md) to learn more.
  
## Roadmap:
I'm working to get this to an MVP state. You can follow the progress in [the project board on Github](https://github.com/nstapelbroek/envoy-swarm-control-plane/projects/1). 