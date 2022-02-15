BUILD ?= $(CURDIR)/bin
IMAGE ?= gnbsim
VERSION ?= v3.1
REMOTE_REGISTRY ?= localhost:5000

TOOLS := $(wildcard ./tools/*)
FLOW := $(wildcard ./flow/*)

export BUILD

.PHONY: all tools flow $(TOOLS) $(FLOW) push clean

all: 
	docker build -f docker/Dockerfile.ubuntu.18.04 -t $(IMAGE):$(VERSION) .

push:
	docker tag $(IMAGE):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)
	docker push $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)

clean:
	rm -rf ./bin
	docker rmi -f $(IMAGE):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)
