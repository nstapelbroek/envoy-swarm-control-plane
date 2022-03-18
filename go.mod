module github.com/nstapelbroek/envoy-swarm-control-plane

go 1.14

require (
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20220314180256-7f1daf1720fc // indirect
	github.com/containerd/containerd v1.6.1 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7
	github.com/envoyproxy/go-control-plane v0.10.1
	github.com/envoyproxy/protoc-gen-validate v0.6.7 // indirect
	github.com/go-acme/lego/v4 v4.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0 // indirect
	github.com/klauspost/compress v1.15.1 // indirect
	github.com/klauspost/cpuid/v2 v2.0.12 // indirect
	github.com/miekg/dns v1.1.47 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/minio-go/v7 v7.0.23
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/rs/xid v1.3.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20220315160706-3147a52a75dd // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220318055525-2edf467146b5 // indirect
	golang.org/x/tools v0.1.10 // indirect
	google.golang.org/genproto v0.0.0-20220317150908-0efb43f6373e // indirect
	google.golang.org/grpc v1.45.0
	gopkg.in/ini.v1 v1.66.4 // indirect
	gotest.tools v2.2.0+incompatible
)

// Overrides from Traefik
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200204220554-5f6d6f3f2203
