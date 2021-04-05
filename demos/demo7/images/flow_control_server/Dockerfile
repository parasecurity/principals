FROM golang

WORKDIR "/home"
COPY tcp_server.go .
RUN go build tcp_server.go

CMD ["./tcp_server"]
