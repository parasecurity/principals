FROM ubuntu:latest

RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install -y libpcap-dev openvswitch-switch

COPY tools/ /home/
