BUILD ?= $(CURDIR)/bin
IMAGE ?= antrea-tsi
VERSION ?= v1.0.5
REMOTE_REGISTRY ?= localhost:5000

TOOLS := $(wildcard ./tools/*)
FLOW := $(wildcard ./flow/*)
STATISTICS := $(wildcard ./statistics/*)

export BUILD

.PHONY: all tools flow statistics $(TOOLS) $(FLOW) $(STATISTICS) push clean

all: tools flow statistics
	docker build . -t $(IMAGE):$(VERSION)

tools: $(TOOLS)

flow: $(FLOW)

statistics: $(STATISTICS)

$(TOOLS):
	mkdir -p ./bin
	$(MAKE) -C $@

$(FLOW):
	mkdir -p ./bin
	$(MAKE) -C $@

$(STATISTICS):
	mkdir -p ./bin
	$(MAKE) -C $@

push:
	docker tag $(IMAGE):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)
	docker push $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)

clean:
	rm -rf ./bin
	docker rmi -f $(IMAGE):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)
