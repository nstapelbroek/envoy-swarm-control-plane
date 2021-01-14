DOCKER_NETWORK_NAME = edge-traffic
ENVOY_SERVICE_NAME = envoy
DOCKER_NETWORK_ID = $(shell docker network ls -q -f name=$(DOCKER_NETWORK_NAME))
USERLAND_STACKS := $(wildcard stacks/*.yml)
DOCKER_BRIDGE_IP := $(shell ip -f inet addr show docker_gwbridge | grep -Po 'inet \K[\d.]+')

ifeq ($(DOCKER_NETWORK_ID),)
	DOCKER_NETWORK_ID := $(shell docker network create --driver=overlay --attachable $(DOCKER_NETWORK_NAME))
endif

.PHONY: cleanup
cleanup:
	$(foreach f, $(USERLAND_STACKS), docker stack rm $(basename $(notdir $f)) || exit 0;)
	@docker service rm $(ENVOY_SERVICE_NAME) || exit 0
	$(info removing network)
	@docker network rm $(DOCKER_NETWORK_NAME) || exit 0

.PHONY: deploy-services
deploy-services:
	$(foreach f, $(USERLAND_STACKS), docker stack deploy --compose-file $(CURDIR)/$f $(basename $(notdir $f));)
	@echo Demo services deployed, make sure you update you /etc/hosts !

.PHONY: deploy-envoy
deploy-envoy:
	@docker image build --tag nstapelbroek/envoy-edge:dev --build-arg CONTROL_PLANE_HOST=$(DOCKER_BRIDGE_IP) $(CURDIR)/envoy
	@docker service create --name $(ENVOY_SERVICE_NAME) --publish published=80,target=80,mode=host --publish published=443,target=443,mode=host --publish published=10000,target=10000,mode=host --network $(DOCKER_NETWORK_NAME) nstapelbroek/envoy-edge:dev

.PHONY: run-pebble
run-pebble:
	@pebble -config $(CURDIR)/pebble/config.json

.PHONY: run
run: deploy-envoy deploy-services run-pebble