FROM alpine:3.15.4

RUN apk add \
    libpcap-dev \
    && rm -rf /var/lib/apt/lists/*

COPY bin/ /home/tsi/bin