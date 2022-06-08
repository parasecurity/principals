.PHONY: all clean image

all: image

clean:
	rm -rf bin/
	docker rmi tsi-api-frontend:v0.0.1 localhost:5000/tsi-api-frontend:v0.0.1

client:
	mkdir -p bin/
	cd api-client && go mod tidy && go build -o ../bin/client cmd/client.go

image: client
	docker build . -t tsi-api-frontend:v0.0.1

push:
	docker tag tsi-api-frontend:v0.0.1 localhost:5000/tsi-api-frontend:v1.0.1
	docker push localhost:5000/tsi-api-frontend:v0.0.1
