.PHONY: all clean image

all: image

clean:
	rm -rf tools/*
	docker rmi tsi-tools:v1.0.0

canary:
	mkdir -p tools/
	go build -o tools/httpCanary httpCanary.go

canary-link:
	mkdir -p tools/
	go build -o tools/httpCanaryLink httpCanaryLink.go

client:
	mkdir -p tools/
	go build -o tools/httpClient httpClient.go

# if build fails run: sudo apt install libpcap-dev
detector:
	mkdir -p tools/
	cd httpDetector && go build -o ../tools/httpDetector httpDetector.go

detector-link:
	mkdir -p tools/
	cd httpDetectorLink && go build -o ../tools/httpDetectorLink httpDetectorLink.go

server:
	mkdir -p tools/
	cd httpServer && go build -o ../tools/httpServer httpServer.go
	cp -r httpServer/static/ tools/

image: canary client detector server canary-link detector-link
	docker build . -t tsi-tools:v1.0.0

push: 
	docker tag tsi-tools:v1.0.0 localhost:5000/tsi-tools:1.0.22
	docker push localhost:5000/tsi-tools:1.0.22
