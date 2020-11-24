# Local development environment

Tools and setup instructions for developing on a local linux machine. It requires a local swarm to run envoy
and any upstream services. The control plane runs on the host to make it easier to debug / develop. External services like 
LetsEncrypt should probably also run outside the swarm so we'll have less complexity around DNS.

To summarise, this setup is packed with:

- The control plane application running on a host
 - XDS server on port 9876
 - ACME Challenges on port 8080, envoy will proxy these requests
- Envoy Proxy running as a swarm service
  - Published port 80, 443 in host mode
  - Admin interface is available at [port 10000](http://localhost:10000)
- [Pebble](https://github.com/letsencrypt/pebble) on the host
  - ACME running on port [14000](https://127.0.0.1:14000/dir)
  - Management running on [15000](https://127.0.0.1:15000/)
  - CA certificate is taken from https://github.com/letsencrypt/pebble/blob/master/test/certs/localhost/cert.pem
- A self-signed certificate for `example.com,www.example.com,new.example.com,frontend.example.com` in the certificates folder
- An example upstream docker swarm stack


## Requirements

- [Golang](https://golang.org/doc/install) for developing. Use at least go 1.13 or up for less headache with gomodules
- [Docker, running as a swarm manager](https://docs.docker.com/engine/reference/commandline/swarm_init/) reachable via `/var/run/docker.sock`
- [Pebble](https://github.com/letsencrypt/pebble#install) for issuing certificates 
- [Make](https://www.gnu.org/software/make/) to automate some setup and teardown
- Manually editing /etc/hosts to point any testing domains to your docker bridge address. In this setup: `172.20.0.1 example.com www.example.com new.example.com frontend.example.com api.example.com`

### Getting started

We'll use make to assure the swarm network exists and to automate building images before starting services with them. 
Since I don't want to send these services
 to a subshell you should pass parallel execution flags (--jobs) to make it a bit more explicit that you are doing multiple things at once.
 An example to start the setup all at once: 
```
make run -j 3
```

You can also run individual parts alone, if you prefer to have multiple terminals open :)
```
make run-envoy
make run-pebble
make deploy-services
```

After running these services, you can run the control plane via your IDE or by hand from the project root:

```bash
LEGO_CA_CERTIFICATES=$(pwd)/deployments/development-host/pebble/ca.pem \
go run $(pwd)/cmd/swarm-control-plane --debug --storage-dir $(pwd)deployments/development-host/certificates/ --acme-email you@provider.com --acme-accept-terms --acme-local
```
