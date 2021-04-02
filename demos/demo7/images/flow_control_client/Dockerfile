FROM golang:alpine3.13

WORKDIR "/home"
COPY tcp_in_client.go tcp_out_client.go /home/
RUN go build tcp_in_client.go
RUN go build tcp_out_client.go

CMD ["./tcp_server"]
