.PHONY: all clean image

all: image

image:
	docker build . -t tsi-dga:v1.0.0

push: 
	docker tag tsi-dga:v1.0.0 localhost:5000/tsi-dga:v1.0.0
	docker push localhost:5000/tsi-dga:v1.0.0

clean:
	docker rmi tsi-api:v1.0.0 localhost:5000/tsi-dga:v1.0.0
