BUILD ?= $(CURDIR)/bin
IMAGE ?= attacks
VERSION ?= v1.0.0
REMOTE_REGISTRY ?= localhost:5000

TOOLS := $(wildcard ./tools/*)


export BUILD

.PHONY: all tools $(TOOLS) push clean

all: tools
	docker build . -t $(IMAGE):$(VERSION)

tools: $(TOOLS)

$(TOOLS):
	mkdir -p ./bin
	$(MAKE) -C $@

push:
	docker tag $(IMAGE):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)
	docker push $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)

clean:
	rm -rf ./bin
	docker rmi -f $(IMAGE):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)
