.PHONY: all clean image

all: image

clean:
	rm -f bin/*
	docker rmi antrea-tsi:v1.0.1

broadcaster:
	mkdir -p bin/
	cd flow/flow-broadcaster && go build -o ../../bin/flow-broadcaster broadcaster.go

server:
	mkdir -p bin/
	cd flow/flow-server && go build -o ../../bin/flow-server server.go

control:
	mkdir -p bin/
	cp flow/flow-control/forward.py bin/

image: broadcaster server control
	docker build . -t antrea-tsi:v1.0.1

push: 
	docker tag antrea-tsi:v1.0.1 localhost:5000/antrea-tsi:v1.0.1
	docker push localhost:5000/antrea-tsi:v1.0.1
	docker rmi localhost:5000/antrea-tsi:v1.0.1