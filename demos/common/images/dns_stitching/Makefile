.PHONY: all clean image

all: image

clean:
	rm -rf bin/
	docker rmi dns-stitching:v0.1.0 localhost:5000/dns-stitching:v0.1.0

image: 
	docker build . -t dns-stitching:v0.1.0

push:
	docker tag dns-stitching:v0.1.0 localhost:5000/dns-stitching:v0.1.0
	docker push localhost:5000/dns-stitching:v0.1.0
