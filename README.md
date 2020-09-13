# Envoy Swarm control plane
Opinionated control plane software that enables Envoy as an edge proxy in a docker swarm cluster.

todo: logo or small demo

todo: add feature list, the problem it solves and what it enables

## Background
Early 2020 I faced a couple of challenges while attempting to isolate my swarm manager nodes from service workloads. The
goal was to increase reliability of the swarm by making it easier to replicate key parts of the infrastructure in 
case of a node failure. Swarm's service mesh would take care of most resiliency workarounds when a replicated part fails. 

Most of my workloads are HTTP services. I figured that replicating the single instance load balancer/proxy to worker  
nodes would help offloading managers and prevent it from being a single point of failure. The challenge here was moving the
state like TLS certificates, container endpoints and LetsEncrypt challenges out of the proxy.
[Envoy](https://envoyproxy.io/) seemed like a good approach for this.
 
As a bonus I wanted to play around with ingress & mesh networks before delegating most of the heavy lifting to CNCF 
products on something bigger like Kubernetes. Docker Swarm presents a chance to keep things simple and cheap for 
the pet projects that I'm running. This project aims to be the same: simple to operate and cheap on the required infrastructure.

### Design scope
- Envoy proxy will take care of receiving traffic on the edge
  - Edge nodes are disposable workers in the swarm
  - You should be able to replicate the proxy for high availability purposes
- We'll use Swarm's routing mesh to route traffic to containers
- The control plane runs on a swarm manager, reading info about service definitions and converting it into the configuration that Envoy understands. 
  
## Roadmap:
I'm working to get this to an MVP state. You can follow the progress in [the project board on Github](https://github.com/nstapelbroek/envoy-swarm-control-plane/projects/1). 