all: server client

server:
	go build -o flow_control_server/tcp_server flow_control_server/tcp_server.go

client:
	go build -o flow_control_client/tcp_in_client flow_control_client/tcp_in_client.go
	go build -o flow_control_client/tcp_out_client flow_control_client/tcp_out_client.go

clean:
	rm -f flow_control_server/tcp_server 
	rm -f flow_control_client/tcp_in_client
	rm -f flow_control_client/tcp_out_client

.PHONY: all server client clean
