# Envoy Swarm control plane
Opinionated control plane software that enables Envoy as an edge proxy in your swarm cluster.

## Context
A recent re-evaluation of my hobby cluster has put me in a twilight zone between Kubernetes and Docker Swarm (again). While
Kube is inevitable I wanted to figure out how things work inside ingress and/or mesh networks before delegating most of the work to a CNCF product.
Docker Swarm presents a chance to keep things simple and relatively cheap. I want a simple (thus opinionated), replicable edge load balancer to pass traffic towards
a handful of swarm services where being available is more important than scaling.

In my old setup I'd use Traefik for this. Being a long-time watcher of certain swarm- or high-availability- challenges in this excellent router I thought I'd try 
a more opinionated setup where I move "the state" out of application. That's when I found Envoy :).

### Design scope
- Envoy proxy will take care of receiving traffic on the edge
  - Edge nodes are disposable workers in the swarm
  - You should be able to replicate the proxy for high availability purposes
- We'll use Swarm's routing mesh to route traffic to containers
  
 ## Plan of approach:
 
 1. The control plane will pull swarm stacks from the management API (docker socket) and convert them into Envoy configuration objects if they are labeled so
 2. Define each labeled service as a cluster (CDS)
 3. Define each exposed port on the labeled service as an Endoint (EDS)
 4. Define each labeled domain on the service as a route (RDS)
 5. Define a set of common filters (connection managers, filters, where to get the SSL certificate, etc.) as a listener (LDS)
 6. Serve TLS certificates for the parsed host using SDS
 
 Keep updating the state either using an interval (Traefik style) or check if we can hook into some docker
 events as we can potentially rely more heavily on the swarm routing mesh.
 
 Then, ship the whole shebang in two seperate containers, one for the control plane and one preconfigured edge router (which is just the envoy with a prebaked yml)
  
 