## Background
I wanted to increase reliability of my swarm setup by making it easier to replicate key parts of the infrastructure in case of a node failure.
As most of my workloads are HTTP services for several domains, I figured that replicating the proxy should prevent it from becoming a single point of failure.
The challenge was moving the state out of the proxy. That state being things route & network configuration, TLS certificates, and the LetsEncrypt tokens used to issue those certificates. 

[Envoy](https://envoyproxy.io/) offers a wide range of configuration options via an API that seemed like a good fit for the job.
 
### Why not use x?
Building this gave me a chance to learn about mesh networks, LetsEncrypt, gRPC, envoy and Docker swarm. 
Instead of delegating most of the work to CNCF products on something like Kubernetes. I used this chance to keep things 
simple (thus opinionated) and cheap for the pet projects that I'm running in the cloud. 