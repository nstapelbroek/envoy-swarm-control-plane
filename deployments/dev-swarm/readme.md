# Local development swarm

## Contains:

- Control and data plane network
- Envoy config for a control plane running at host.docker.internal (xds port 9876, acme port 8080)
- 2 Upstream swarm service clusters
    - Route config on one domain, two paths: / and /api
    - 1 Matching TLS certificate if you point the storage towards ./certificates
- Letsencrypt test setup
    - Pebble configuration and CA certificate
- Cleanup and deploy scripts you can run from the project root

## Assumes:

- Control plane runs on host
- Pebble runs on host

## Running this:

Assuming that you are in the project root, you can run the setup by:

1. Starting the swarm stacks:

```bash
$(pwd)/deployments/dev-swarm/deploy.sh
```

1. Running the app:

```bash
LEGO_CA_CERTIFICATES=$(pwd)/deployments/dev-swarm/certificates/pebble.pem \
go run $(PWD)/cmd/swarm-control-plane --debug --storage-dir $(pwd)/deployments/dev-swarm/certificates/ --acme-email you@provider.com --acme-accept-terms
```
