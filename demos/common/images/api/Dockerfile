FROM ubuntu:20.04

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y \
    wget \
    curl \
    vim \
    golang \
    apt-transport-https \
    ca-certificates \
    sudo \
    tmux \
    gnupg2

RUN mkdir -p /etc/apt/keyrings

RUN echo "deb [signed-by=/etc/apt/keyrings/kubernetes.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
RUN curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes.gpg

RUN apt-get update && apt-get -y install kubectl

COPY bin/server /home/
COPY bin/client /home/
COPY api/internal /home/internal/

WORKDIR /home

