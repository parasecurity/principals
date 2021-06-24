.PHONY: all clean image

all: image

image:
	docker build . -t tsi-honeypot:v1.0.0

push: 
	docker tag tsi-honeypot:v1.0.0 localhost:5000/tsi-honeypot:v1.0.0
	docker push localhost:5000/tsi-honeypot:v1.0.0

clean:
	docker rmi tsi-honeypot:v1.0.0 localhost:5000/tsi-honeypot:v1.0.0
