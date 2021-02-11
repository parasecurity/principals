#!/usr/bin/env bash
set -euo pipefail

if [ "$EUID" -ne 0 ]; then
        echo "Please run as root"
        exit
fi

if [[ "$(which minikube)" != "" ]]; then
        echo "$(minikube version --short)"
        exit
fi

curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube_latest_amd64.deb
dpkg -i minikube_latest_amd64.deb
rm minikube_latest_amd64.deb

chmod 755 /usr/bin/minikube
mkdir -p $HOME/.minikube
chown -R $SUDO_USER $HOME/.minikube; chmod -R u+wrx $HOME/.minikube

minikube completion bash | tee -a /etc/bash_completion.d/minikube

echo "minikube installed successfully"
echo "Please logout and login for changes to take effect"
