FROM projects.registry.vmware.com/antrea/antrea-ubuntu:v1.0.0

RUN apt-get update && apt-get install -y \
    iperf3 \
    libpcap-dev \
    sudo \
    nodejs \
 && rm -rf /var/lib/apt/lists/*

COPY bin/ /home/tsi/bin
COPY scripts/ /home/tsi/scripts
