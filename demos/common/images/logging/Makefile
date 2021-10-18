.PHONY: all clean image

all: server agent client

clean:
	rm -rf bin/
	# docker rmi tsi-api:v1.0.0 localhost:5000/tsi-logging:v1.0.0

server:
	mkdir -p bin/
	go build -o bin/server cmd/log-server/server.go

agent:
	mkdir -p bin/
	go build -o bin/agent cmd/log-agent/agent.go

client:
	mkdir -p bin/
	go build -o bin/client cmd/client/client.go
	go build -o bin/multithread_client cmd/multithread_client/multithread_client.go

image: server agent client
	docker build . -t tsi-logging:v1.0.0

push:
	docker tag tsi-logging:v1.0.0 localhost:5000/tsi-logging:v1.0.0
	docker push localhost:5000/tsi-logging:v1.0.0
