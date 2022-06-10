FROM alpine:3.15.4

# gcompat is used in order to execute binaries
RUN apk add \
    libpcap-dev \
    gcompat \
    && rm -rf /var/lib/apt/lists/*

COPY bin/ /home/tsi/bin