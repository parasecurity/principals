VERSION ?= v1.0.2
IMAGE ?= tsi-logging
REMOTE_REGISTRY ?= localhost:5000

.PHONY: all clean image

all: server agent client

clean:
	rm -rf bin/ 
	docker rmi $(IMAGE):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)

server:
	mkdir -p bin/
	go build -o bin/server cmd/log-server/server.go cmd/log-server/parser.go cmd/log-server/sorter.go

agent:
	mkdir -p bin/
	go build -o bin/agent cmd/log-agent/agent.go

client:
	mkdir -p bin/
	go build -o bin/client cmd/client/client.go
	go build -o bin/multithread_client cmd/multithread_client/multithread_client.go

image: server agent client
	docker build . -t $(IMAGE):$(VERSION)

push: image
	docker tag $(IMAGE):$(VERSION) $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)
	docker push $(REMOTE_REGISTRY)/$(IMAGE):$(VERSION)
