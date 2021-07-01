FROM projects.registry.vmware.com/antrea/antrea-ubuntu:v1.0.0

COPY bin/ /home/tsi/bin
COPY scripts/ /home/tsi/scripts

RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install -y libpcap-dev iperf3