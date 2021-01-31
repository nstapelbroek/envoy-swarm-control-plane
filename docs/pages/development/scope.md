## Design scope
- Envoy proxy will take care of receiving traffic on the edge
    - You should be able to replicate the proxy for high availability purposes
- We'll use Swarm's routing mesh to route traffic to containers
  - Using DNS queries to docker's internal `tasks.` name, so we are left for future options around balancing and health checks
- The control plane runs on a swarm manager
  - Control plane should only read from the socket, no need to write as this creates too much responsibility in managing your swarm
  - There is no need to expose the control plane to the internet. Things like LetsEncrypt should also proxy through the Envoy instances.

## Limitations
Current decisions that I made to cut the scope a bit:

- Only one endpoint per service
  - Use TCP for communication to services
- Initially, only build this to route HTTP traffic on port 80 and 443
- HTTPs is always with redirect
  - When enabled, LetsEncrypt will always issue a certificate. No overriding this per service
