## What is this? 
Envoy Swarm Control Plane is a piece of software that's designed to run inside a Docker Swarm environment.
It will read swarm service definitions and convert them to configuration objects that Envoy proxies can use.

The goal is to run envoy at the edge as a reverse proxy for your swarm services.
this means that the focus is on serving external clients rather than facilitating in service-to-service communication. 

## Installation
Below are the steps you should take to set up and run this software. I keep the most up-to-date code examples in the 
[deployments](https://github.com/nstapelbroek/envoy-swarm-control-plane/tree/master/deployments) folder of this repository.
Note that this folder might contain some files and folders that are not relevant for these instructions.

### 1. Setting up your docker swarm 
![Docker Swarm Rocks](https://dockerswarm.rocks/img/logo-light-blue-vectors.svg)

If you are new to docker swarm or do not have a cluster right now: try the tutorial at https://dockerswarm.rocks/ to get you started.
This control plane works on all kinds of setups regardless if you have 1,3 or 10+ nodes running. 

### 2. Creating a network

The envoy instances need a network to route requests on towards your services. We'll use a [overlay network](https://docs.docker.com/network/overlay/) 
specifically to communicate between the reverse proxy and any container that should handle requests from the internet.

Here's a command to get you started, feel free to tweak it to your needs. 

```
docker network create --driver=overlay --attachable edge-traffic
```

Keep in mind the overlay network name that you've just created as you will be referring to this network in your docker-compose stacks.

### 3. Decide how the internet connects to your cluster 

todo: write about where you DNS points towards and how you should prepare these hosts with labels.

### 4. Deploy the stack 

todo: copy paste usable stack from deployments folder


