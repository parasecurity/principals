FROM ubuntu:20.04

RUN apt-get update && apt-get install -y \
    iperf3 \
    tcpdump \
    libpcap-dev \
 && rm -rf /var/lib/apt/lists/*

COPY bin/ /home/tsi/bin
#COPY scripts/ /home/tsi/scripts
