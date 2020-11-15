# Local development environment

Tools and setup instructions for developing on a local unix / macos machine. We're going
to run almost exclusively on the host except for some demo swarm services.

Furthermore, this setup is packed/opinionated with:

- A self-signed certificate for `example.com,www.example.com,new.example.com,frontend.example.com`
- Pebble config with CA certificate and Key, copied from https://github.com/letsencrypt/pebble/blob/master/test/certs/localhost/cert.pem
- Envoy configuration that will call "a control plane" at localhost:9000

## Requirements

- [Golang](https://golang.org/doc/install) for developing. Use at least go 1.13 or up for less headache with gomodules
- [Docker, running as a swarm manager](https://docs.docker.com/engine/reference/commandline/swarm_init/) reachable via `/var/run/docker.sock` 
- [Pebble](https://github.com/letsencrypt/pebble#install) for issuing certificates
- [GetEnvoy](https://www.getenvoy.io/install/) to run the Envoy proxy
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
