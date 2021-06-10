.PHONY: all clean image

all: image

clean:
	rm -rf bin/
	docker rmi antrea-tsi:v1.0.0

broadcaster:
	mkdir -p bin/
	cd flow/flow-broadcaster && go build -o ../../bin/flow-broadcaster broadcaster.go

server:
	mkdir -p bin/
	cd flow/flow-server && go build -o ../../bin/flow-server server.go

control:
	mkdir -p bin/
	cp flow/flow-control/forward.py bin/

detector:
	mkdir -p bin/
	cd tools/httpDetector && go build -o ../../bin/detector httpDetector.go

detectorLink:
	mkdir -p bin/
	cd tools/httpDetectorLink && go build -o ../../bin/detectorLink httpDetectorLink.go

canary:
	mkdir -p bin/
	cd tools/httpCanary && go build -o ../../bin/canary httpCanary.go

canaryLink:
	mkdir -p bin/
	cd tools/httpCanaryLink && go build -o ../../bin/canaryLink httpCanaryLink.go

client:
	mkdir -p bin/
	cd tools/httpClient && go build -o ../../bin/client httpClient.go

site:
	mkdir -p bin/
	cd tools/httpServer && go build -o ../../bin/site httpServer.go

image: broadcaster server control detector detectorLink canary canaryLink client site
	docker build . -t antrea-tsi:v1.0.0

push: 
	docker tag antrea-tsi:v1.0.0 localhost:5000/antrea-tsi:v1.0.0
	docker push localhost:5000/antrea-tsi:v1.0.0
