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
    tmux

RUN curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] http://packages.cloud.google.com/apt/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list && \
    apt-get update && apt-get -y install kubectl

COPY bin/server /home/
COPY bin/client /home/
COPY api/internal /home/internal/

WORKDIR /home
