## Installation

Below are the steps you should take to set up and run this software. I keep the most up-to-date code examples in the
[deployments](https://github.com/nstapelbroek/envoy-swarm-control-plane/tree/master/deployments) folder of this
repository. Note that this folder might contain some files and folders that are not relevant for these instructions.

### What we'll be configuring

This setup guide works regardless if you have 3, 5 or 10 hosts in your swarm. This is because we'll address the nodes by
their role instead of their name. Please remember which node plays what kind of role when following the guide.

A high-over overview for our setup consists of:

1. Your DNS servers are pointing towards the ede nodes for the domains you want to host (out of scope for this setup guide)
1. One or many nodes who are labeled as `edge` nodes. These nodes will:
    1. Run Envoy Instances
    1. Accept traffic from the internet on port 80 and 443
    1. Use [host mode](https://docs.docker.com/network/host/) for exposing ports to prevent unnecessary network hops.
1. One swarm manager node that runs the control plane software.


A diagram of this setup is given below. You'll notice that you can mix & match the roles to get to a setup that you
prefer. For example, you can handle incoming HTTP traffic on swarm worker nodes and keep your managers isolated!

// todo diagram

### 1. Setting up your docker swarm

![Docker Swarm Rocks](https://dockerswarm.rocks/img/logo-light-blue-vectors.svg)

If you are new to docker swarm or do not have a cluster at this moment: try the steps described
at https://dockerswarm.rocks/ to get you started. When your cluster is all setup, connect to a manager node and proceed.

### 2. Creating a network

The envoy instances need a network to communicate with your upstream services. We'll use
a [overlay network](https://docs.docker.com/network/overlay/)
specifically to communicate between the reverse proxy and any container that you've configured to handle requests from
the internet.

Here's a command that works for most setups. Feel to tweak it if you prefer a certain subnet mask or encryption.

```bash
docker network create --driver=overlay --attachable edge-traffic
```

Keep in mind the overlay network name (`edge-traffic`) that you've just created as you will be referring to this network
in your stack files.

### 3. Label your edge nodes

We are going to label one or more nodes in your cluster as an edge node. Labeling the node allows the docker swarm
scheduler to filter for specific nodes when deploying the envoy application.

In the example below I have a pair of nodes in a manager & worker setup. I'm going to label the worker node as an edge
node.

![example where I label a node](../assets/node_labeling.png)

```sh
docker node update --label-add edge=true  $NODE_ID
```

Of course, you will have to point the DNS entries for your domain towards the public IP address of the node.

#### A note on using multiple nodes

When using multiple edge nodes you would also have to think about how internet traffic is divided between these nodes.
It's a bit more complicated than adding another IP address to the DNS record.

### 4. Deploy the envoy stack

Here is the stack file you can use to get started instantly. It refers to the network we've created in step 2 and
deploys the envoy instances on the edge nodes from step 3.

```yaml
version: '3.7'

services:
  control-plane:
    image: nstapelbroek/envoy-swarm-control-plane:0.1
    command:
      - --ingress-network
      - edge-traffic
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager
      restart_policy:
        condition: any
        window: 10s
    networks:
      - default
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro

  proxy:
    image: nstapelbroek/envoy-swarm-edge:0.1
    deploy:
      mode: global
      placement:
        constraints:
          - node.labels.edge == true
      restart_policy:
        condition: any
        window: 10s
    networks:
      - default
      - edge-traffic
    ports:
      - target: 80
        published: 80
        mode: host
      - target: 443
        published: 443
        mode: host


networks:
  default: { }
  edge-traffic:
    external: true
```

To deploy this, save the config to a file e.g. stack.yml and deploy it like so:

```sh
docker stack deploy --compose-file stack.yml envoy
```

Then, wait a couple of seconds to let docker pull the images, schedule the tasks and deploy the containers. You can
check the status with:

```sh
docker stack ps envoy
```

The output should be something like this:

```
ID                  NAME                                     IMAGE                                        NODE                DESIRED STATE       CURRENT STATE          ERROR                         PORTS
m6l8ca4jqc4c        envoy_proxy.vioymhujto4sv1bqlw7d20wcm    nstapelbroek/envoy-swarm-edge:0.1            worker01            Running             Running 5 weeks ago                                  *:443->443/tcp,*:80->80/tcp   
o9zzxc1bnokt        envoy_control-plane.1                    nstapelbroek/envoy-swarm-control-plane:0.1   manager01           Running             Running 4 weeks ago                                  
```

When the services are running, the control plane will read your docker swarm state and communicate this towards the
proxies.

#### Encrypting the web

I highly recommend setting up TLS and optional LetsEncrypt. These are just a couple of extra command arguments for your
control plane. Read about it here.

// todo write TLS setup docs

### 5. Add labels to your services

The reverse proxy is running and ready to route traffic. We now have to update or create our services with the right
labels, so the control plane knows which service configuration should reside in the Envoy proxies.

A bare minimal configuration includes a port where your container accepts HTTP traffic and a domain / hostname. An
example:

```yaml
version: '3.7'

services:
  frontend:
    image: nstapelbroek/static-webserver:3
    deploy:
      labels:
        - envoy.endpoint.port=80
        - envoy.route.domain=example.com
    networks:
      - edge-traffic

networks:
  edge-traffic:
    external: true
```

After deploying your application stack / services with the labels. The Envoy proxy will be able to route traffic towards
them. It works just like [in the demo](https://asciinema.org/a/LEu3l3sLfIVVA6GomAh5cn0Mo).

#### There is more

There are more configuration labels available! you can route paths, multiple domains and even things like connection
timeouts. See a list of more options here.

// todo write label config docs

