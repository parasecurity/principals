BUILD ?= $(CURDIR)/bin
IMAGE ?= tls-fingerprint
VERSION ?= v0.1.0
REMOTE_REGISTRY ?= localhost:5000

.PHONY: all tool push clean

all: 
	docker build . -t $(IMAGE):$(VERSION)

tool: 
	mkdir -p ./bin
	go build -o $(BUILD)/$(IMAGE)

push:
	docker tag $(IMAGE):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)
	docker push $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)

clean:
	rm -rf ./bin
	docker rmi -f $(IMAGE):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)