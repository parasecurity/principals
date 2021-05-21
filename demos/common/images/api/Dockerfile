FROM ubuntu:20.04

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && \
    apt-get upgrade && \
    apt-get -y install \
    wget \
    curl \
    vim \
    golang \
    apt-transport-https \
    ca-certificates \
    sudo \
    tmux

RUN curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list && \
    apt-get update && apt-get -y install kubectl

COPY go/api /home/api
COPY api-client/ /home/api-client

WORKDIR /home
