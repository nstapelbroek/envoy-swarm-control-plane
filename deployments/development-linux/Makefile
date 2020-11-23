DOCKER_NETWORK_NAME = edge-traffic
ENVOY_SERVICE_NAME = envoy
DOCKER_NETWORK_ID = $(shell docker network ls -q -f name=$(DOCKER_NETWORK_NAME))
USERLAND_STACKS := $(wildcard stacks/*.yml)
ENVOY_VERSION = 1.15.1
ifeq ($(DOCKER_NETWORK_ID),)
	DOCKER_NETWORK_ID := $(shell docker network create --driver=overlay --attachable $(DOCKER_NETWORK_NAME))
endif

.PHONY: cleanup
cleanup:
	$(info removing demo services)
	$(foreach f, $(USERLAND_STACKS), @docker stack rm $(basename $(notdir $f));)
	$(info removing pebble)
	@docker service rm pebble
	$(info removing envoy)
	@docker service rm enovy
	$(info removing network)
	@docker network rm $(DOCKER_NETWORK_NAME)

.PHONY: deploy-services
deploy-services:
	$(foreach f, $(USERLAND_STACKS), @docker stack deploy --compose-file $(CURDIR)/$f $(basename $(notdir $f));)
	@echo Demo services deployed

.PHONY: run-envoy
run-envoy:
	@docker image build --tag nstapelbroek/envoy-edge:dev $(CURDIR)/envoy
	@docker service create --name envoy --publish published=80,target=80,mode=host --publish published=443,target=443,mode=host --publish published=10000,target=10000,mode=host --network $(DOCKER_NETWORK_NAME) nstapelbroek/envoy-edge:dev

.PHONY: run-pebble
run-pebble:
	@docker image build --tag nstapelbroek/pebble:dev $(CURDIR)/pebble
	@docker service create --name pebble --publish 14000:14000 --publish 15000:14000 --network $(DOCKER_NETWORK_NAME) nstapelbroek/pebble:dev

.PHONY: run
run: run-envoy run-pebble deploy-services