.PHONY: all clean image

all: image

clean:
	docker rmi tsi-api-client:v0.0.1 localhost:5000/tsi-api-client:v0.0.1

image: 
	docker build . -t tsi-api-client:v0.0.1

push:
	docker tag tsi-api-client:v0.0.1 localhost:5000/tsi-api-client:v0.0.1
	docker push localhost:5000/tsi-api-client:v0.0.1
