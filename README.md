# Envoy Swarm control plane
Opinionated control plane software that enables Envoy as an edge proxy in your swarm cluster.

## Context
A recent re-evaluation of my hobby cluster has put me in a twilight zone between Kubernetes and Docker Swarm (again). While
Kube is inevitable I wanted to figure out how things work inside ingress and/or mesh networks before delegating most of the work to a CNCF product.
Docker Swarm presents a chance to keep things simple and relatively cheap. This matches my use case as I only have run a handfull of stacks that don't need to scale but need to be "available".

### Design scope
- Envoy proxy will take care of receiving traffic on the edge
  - Edge nodes are disposable workers in the swarm
  - You should be able to replicate the proxy for high availability purposes
- We'll use swarms ingress network to route traffic to containers
  - Using the VIP or DNS resolution to keep things simple
  
  
 