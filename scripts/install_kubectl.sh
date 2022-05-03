#!/usr/bin/env bash
set -euo pipefail

if [ "$EUID" -ne 0 ]; then
        echo "Please run as root"
        exit
fi

if [[ "$(which kubectl)" != "" ]]; then
        echo "$(kubectl version --short)"
        exit
fi

apt -y install \
        apt-transport-https \
        curl \
        conntrack \
        gnupg2

curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -

echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | \
        tee -a /etc/apt/sources.list.d/kubernetes.list

apt update
apt install -y kubectl

kubectl completion bash | tee -a /etc/bash_completion.d/kubectl

echo "kubectl installed successfully"
echo "Please logout and login for changes to take effect"
