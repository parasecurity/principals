BUILD ?= $(CURDIR)/bin
IMAGE ?= antrea-tsi
VERSION ?= v1.0.0
REMOTE_REGISTRY ?= localhost:5000

TOOLS := $(wildcard ./tools/*)
FLOW := $(wildcard ./flow/*)

export BUILD

.PHONY: all tools flow $(TOOLS) $(FLOW) push clean

all: tools flow
	docker build . -t $(IMAGE):$(VERSION)

tools: $(TOOLS)

flow: $(FLOW)

$(TOOLS):
	mkdir -p ./bin
	$(MAKE) -C $@

$(FLOW):
	mkdir -p ./bin
	$(MAKE) -C $@

push:
	docker tag $(IMAGE):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)
	docker push $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)

clean:
	rm -rf ./bin
	docker rmi -f $(IMAGE):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)
