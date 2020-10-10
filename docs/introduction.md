## What is this? 
Envoy Swarm Control Plane is a piece of software that's designed to run inside a Docker Swarm environment.
It will read swarm service definitions and convert them to configuration objects that Envoy proxies can use.

It's designed with the intent to run envoy at the edge, meaning it's focus is more on serving external clients rather
than facilitating in service-to-service communication. 

## Getting started

todo: better docs once I actually have an MVP out

1. Run swarm
    1. Assumed that you have at least 1 worker and 1 manager
    1. Assumed that the worker is reachable from the internet (DNS records etc.)
1. Use a stackfile from the deployments folder (todo, currently only has the dev setup)
    1. Change the stackfile to fit your needs (acme-email etc.)
1. Deploy

