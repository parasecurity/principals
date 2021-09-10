FROM ubuntu:20.04

ENV DEBIAN_FRONTEND noninteractive 
ENV NETWORK_INTERFACE eth0

RUN apt-get update && apt-get -y install \
    wget \
    build-essential \
    gcc \
    libpcre3-dev \
    zlib1g-dev \
    libluajit-5.1-dev \
    libpcap-dev \
    openssl \
    libssl-dev \
    libnghttp2-dev \
    libdumbnet-dev \
    bison \
    flex \
    libdnet \
    autoconf \
    libtool

WORKDIR /opt

ENV DAQ_VERSION 2.0.7
RUN wget https://www.snort.org/downloads/snort/daq-${DAQ_VERSION}.tar.gz \
    && tar xvfz daq-${DAQ_VERSION}.tar.gz \
    && cd daq-${DAQ_VERSION} \
    && ./configure; make; make install

ENV SNORT_VERSION 2.9.18.1
RUN wget https://www.snort.org/downloads/snort/snort-${SNORT_VERSION}.tar.gz \
    && tar xvfz snort-${SNORT_VERSION}.tar.gz \
    && ls \
    && cd snort-${SNORT_VERSION} \
    && ./configure --enable-sourcefire; make; make install

RUN ldconfig
RUN mkdir -p /etc/snort/rules
ADD rules/local.rules /etc/snort/rules

RUN apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
    && rm -rf /opt/snort-${SNORT_VERSION}* /opt/daq-${DAQ_VERSION}*
