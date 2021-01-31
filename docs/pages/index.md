# Envoy Swarm Control Plane

## What's this?

Envoy Swarm Control Plane is a piece of software that's designed to run inside a Docker Swarm environment. It will read
swarm service definitions and convert them to configuration objects that Envoy proxies can use.

> Envoy Proxy is a modern, high performance, small footprint edge and service proxy. Envoy is most comparable to software load balancers such as NGINX and HAProxy.

source: [Ambassador](https://www.getambassador.io/learn/kubernetes-glossary/envoy-proxy/#what-is-envoy-proxy)

The control plane will configure Envoy to route traffic from the internet towards the many services running on one or
more hosts in your swarm. Envoy itself can do much more e.g. handle service-to-service communication, but with this
control plane we'll focus purely on serving external clients.

### Why do I need this?

Docker swarm does not come with any HTTP routing out of the box. If you are running multiple web services on your swarm
you'll quickly find yourself in a conflict as only one service can be published on the HTTP or HTTPS ports. This is a
common problem usually solved by introducing
a [reverse proxy](https://www.cloudflare.com/learning/cdn/glossary/reverse-proxy/) that takes care of routing traffic
and doing other helpful things like terminating TLS.

As you can imagine, these reverse proxies need to be configured. In an environment where addresses change frequently
e.g. Docker containers starting and stopping, you might need to reconfigure the proxy multiple times a day. That's
something we are automating with this control plane software!

### Not the first in its kind

There are many good solutions for this problem already. We'll focus a bit on the differences here. 
// todo: create comparison and link the word "here".

### What's next?

Ready to give it a try?

[View installation docs](./getting-started/installation.md){: .md-button .md-button--primary }