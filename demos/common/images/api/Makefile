.PHONY: all clean image

all: image

clean:
	rm -rf bin/
	docker rmi tsi-api:v1.0.0 localhost:5000/tsi-api:v1.0.0

server:
	mkdir -p bin/
	cd api && go build -o ../bin/server cmd/server/server.go

client:
	mkdir -p bin/
	cd api-client && go build -o ../bin/client cmd/client.go

image: server client
	docker build . -t tsi-api:v1.0.0

push:
	docker tag tsi-api:v1.0.0 localhost:5000/tsi-api:v1.0.0
	docker push localhost:5000/tsi-api:v1.0.0
