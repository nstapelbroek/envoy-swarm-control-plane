DOCKER_NETWORK_NAME = edge-traffic
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
	$(info removing network)
	@docker network rm $(DOCKER_NETWORK_NAME)

.PHONY: deploy-services
deploy-services:
	$(foreach f, $(USERLAND_STACKS), @docker stack deploy --compose-file $(CURDIR)/$f $(basename $(notdir $f));)
	@echo Demo services deployed

.PHONY: run-envoy
run-envoy:
	@sudo getenvoy run standard:$(ENVOY_VERSION) -- --config-path $(CURDIR)/envoy/config.yaml --log-level debug

.PHONY: run-pebble
run-pebble:
	@pebble -config $(CURDIR)/pebble/config.json

.PHONY: run
run: deploy-services run-envoy run-pebble