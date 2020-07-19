# Envoy Swarm control plane
Opinionated control plane software that enables Envoy as an edge proxy in your swarm cluster.

## Context
A recent re-evaluation of my cluster has put me in a twilight zone between Kubernetes and Docker Swarm (again). While
Kubernetes is inevitable I wanted to play around with ingress & mesh networks before delegating most of the heavy lifting to CNCF products.
Docker Swarm presents a chance to keep things simple and relatively cheap. I wanted an easy configurable (thus opinionated),
replicable edge proxy to pass HTTP traffic to a handful of swarm services using its build-in routing mesh.

In my former setup I used [Traefik](https://traefik.io). Being a long-time watcher of certain swarm- or high-availability- challenges in this excellent router I thought I'd try 
a more opinionated setup where less state resides in the Router itself. [Envoy](https://envoyproxy.io/) seemed like a good tool for this, so here we are.

### Design scope
- Envoy proxy will take care of receiving traffic on the edge
  - Edge nodes are disposable workers in the swarm
  - You should be able to replicate the proxy for high availability purposes
- We'll use Swarm's routing mesh to route traffic to containers
- The control plane runs on a swarm manager, reading info about service definitions and converting it into the configuration that Envoy understands. 
  
## Roadmap:
I'm working to get this to an MVP state. You can follow the progress in [the project board on Github](https://github.com/nstapelbroek/envoy-swarm-control-plane/projects/1). 