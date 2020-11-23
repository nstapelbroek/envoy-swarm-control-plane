# Local development environment

Tools and setup instructions for developing on a local linux machine. We're going
to run almost everything in containers to simulate a production setup as close as possible.

Furthermore, this setup is packed with:

- Envoy 15
  - Configured to run the Admin interface at localhost:10000
  - Configured to call an XDS server at localhost:9876
  - Configured to proxy pass ACME challenges to localhost:8080
- [Pebble](https://github.com/letsencrypt/pebble)
  - ACME running on port [14000](https://0.0.0.0:14000/dir)
  - Management running on [15000](https://0.0.0.0:15000/)
  - CA certificate is taken from https://github.com/letsencrypt/pebble/blob/master/test/certs/localhost/cert.pem
- A self-signed certificate for `example.com,www.example.com,new.example.com,frontend.example.com`


## Requirements

- [Golang](https://golang.org/doc/install) for developing. Use at least go 1.13 or up for less headache with gomodules
- [Docker, running as a swarm manager](https://docs.docker.com/engine/reference/commandline/swarm_init/) reachable via `/var/run/docker.sock` 
- [Make](https://www.gnu.org/software/make/) to automate some setup and teardown  

### Getting started

We use make as a simple task runner that starts and configures our local services. Since I don't want to send these services
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
