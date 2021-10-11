.PHONY: all clean image

all: server agent client

clean:
	rm -rf bin/
	# docker rmi tsi-api:v1.0.0 localhost:5000/tsi-api:v1.0.0

server:
	mkdir -p bin/
	go build -o bin/server cmd/log-server/server.go

agent:
	mkdir -p bin/
	go build -o bin/agent cmd/log-agent/agent.go

client:
	mkdir -p bin/
	go build -o bin/client cmd/client/client.go

# image: server client
# 	./api/scripts/generateCert/generateCert.sh
# 	docker build . -t tsi-api:v1.0.0

# push:
# 	docker tag tsi-api:v1.0.0 localhost:5000/tsi-api:v1.0.0
# 	docker push localhost:5000/tsi-api:v1.0.0
