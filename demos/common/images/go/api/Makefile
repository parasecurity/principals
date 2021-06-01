.PHONY: all clean run_server run_client

all: mtd_client mtd_server

clean:
	rm -f bin/*

gencert: 
	

api_client: gencert
	mkdir -p bin/api-client
	cd api-client && go build -o ../bin/api-client/api-client cmd/api-client/client.go

api_server: gencert
	mkdir -p bin/api-server
	cd api-server && go build -o ../bin/api-server/api-server cmd/api-server/server.go
