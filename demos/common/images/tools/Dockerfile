FROM golang:1.14.7

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && apt-get -y install \
    wget \
    vim \
    tcpdump \
    libpcap-dev

WORKDIR /home

COPY httpCanary.go httpDetector.go httpClient.go /home/
RUN go get -d github.com/google/gopacket && \
    go build httpCanary.go && \
    go build httpDetector.go && \
    go build httpClient.go && \
    rm -rf httpCanary.go httpDetector.go httpClient.go
