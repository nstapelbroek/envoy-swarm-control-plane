# The build environment
FROM --platform=$BUILDPLATFORM golang:1-alpine as base

# Wrap /go/bin/go with a script that converts TARGETPLATFORM to GOARCH format
COPY --from=tonistiigi/xx:golang / /
ARG TARGETPLATFORM
ENV GO111MODULE=on

RUN apk add --no-cache musl-dev git gcc

ADD . /src
WORKDIR /src

# any stage based upon base can be executed in parallel
FROM base as build
RUN cd cmd/swarm-control-plane && go env && go build -v

FROM base as codestyle
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.29.0
RUN golangci-lint run ./...

FROM base as tests
RUN go test ./...

# the actuall result should be concise
FROM alpine:latest as app
RUN apk add -U ca-certificates && rm -rf /var/cache/apk/*
RUN mkdir -p /etc/ssl/certs/le
WORKDIR /bin/
COPY --from=build /src/cmd/swarm-control-plane/swarm-control-plane .

EXPOSE 9876
ENTRYPOINT ["/bin/swarm-control-plane"]