.PHONY: all clean image

all: image

image:
	docker build . -t tsi-snort:v1.0.0

push: 
	docker tag tsi-snort:v1.0.0 localhost:5000/tsi-snort:v1.0.0
	docker push localhost:5000/tsi-snort:v1.0.0

clean:
	docker rmi tsi-snort:v1.0.0 localhost:5000/tsi-snort:v1.0.0
