# Envoy Swarm control plane
Opinionated control plane software that enables Envoy as an edge proxy in a docker swarm cluster.

todo: logo or small demo

todo: add feature list, the problem it solves and what it enables

### Design scope
- Envoy proxy will take care of receiving traffic on the edge
  - Edge nodes are disposable workers in the swarm
  - You should be able to replicate the proxy for high availability purposes
- We'll use Swarm's routing mesh to route traffic to containers
- The control plane runs on a swarm manager, reading info about service definitions and converting it into the configuration that Envoy understands. 
  
## Roadmap:
I'm working to get this to an MVP state. You can follow the progress in [the project board on Github](https://github.com/nstapelbroek/envoy-swarm-control-plane/projects/1). 