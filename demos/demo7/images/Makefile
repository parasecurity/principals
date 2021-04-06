all: server

server:
	go build -o flow_control_server/tcp_server flow_control_server/tcp_server.go

clean:
	rm -f flow_control_server/tcp_server 

.PHONY: all server clean
