.PHONY: all clean image

all: image

clean:
	rm -rf bin/
	docker rmi tsi-api:v1.0.1 localhost:5000/tsi-api:v1.0.1

server:
	mkdir -p bin/
	cd api && go mod tidy && go build -o ../bin/server cmd/server/server.go

client:
	mkdir -p bin/
	cd api-client && go mod tidy && go build -o ../bin/client cmd/client.go

image: server client
	./api/scripts/generateCert/generateCert.sh
	docker build . -t tsi-api:v1.0.1

push:
	docker tag tsi-api:v1.0.1 localhost:5000/tsi-api:v1.0.1
	docker push localhost:5000/tsi-api:v1.0.1
