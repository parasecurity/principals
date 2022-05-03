FROM golang:1.16-alpine 

RUN mkdir -p /usr/local/go/src/dns_stitching
COPY  . /usr/local/go/src/dns_stitching/
WORKDIR /usr/local/go/src/dns_stitching/

RUN apk update
RUN apk add build-base
RUN apk add libpcap-dev
RUN apk add --update --no-cache python3 && ln -sf python3 /usr/bin/python
RUN apk add bash

RUN go mod download
RUN go build -o "/app/dns_stitcher"

WORKDIR /usr/local/go/src/dns_stitching/

ENTRYPOINT ./entrypoint.sh

