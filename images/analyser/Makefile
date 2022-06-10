.PHONY: all clean image

all: image

image:
	docker build . -t tsi-analyser:v1.0.0

push: 
	docker tag tsi-analyser:v1.0.0 localhost:5000/tsi-analyser:v1.0.0
	docker push localhost:5000/tsi-analyser:v1.0.0

clean:
	docker rmi tsi-analyser:v1.0.0 localhost:5000/tsi-analyser:v1.0.0
