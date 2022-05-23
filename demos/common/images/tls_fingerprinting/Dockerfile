FROM golang:1.16-alpine 

COPY  . /home/tsi/
WORKDIR /home/tsi/

RUN apk update
RUN apk add build-base
RUN apk add libpcap-dev

RUN go mod download
RUN go build -o "./bin/tls-fingerprint"