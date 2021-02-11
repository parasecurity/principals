# Setup Kubernetes, Docker on phobos3

For some kubernetes related operations, a browser is needed. Currently, only chromium
have been tested successfully in local vm example. It might have latency issues when
installed in phobos3, so we have not yet addded it there until we really need it.

## Install docker

Use the following script:

```sh
#!/usr/bin/env bash
set -euo pipefail

if [ "$EUID" -ne 0 ]; then
	echo "Please run as root"
	exit
fi

if [[ "$(which docker)" != "" ]]; then
	echo "$(docker -v)"
	exit
fi

apt -y install \
	apt-transport-https \
	ca-certificates \
	curl \
	gnupg-agent \
	software-properties-common

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -

apt-key fingerprint 0EBFCD88

add-apt-repository \
	"deb [arch=amd64] https://download.docker.com/linux/ubuntu \
	$(lsb_release -cs) \
	stable"

apt update
apt -y install \
	docker-ce \
	docker-ce-cli \
	containerd.io

usermod -aG docker "$SUDO_USER"

echo "docker installed successfully"
echo "Please logout and login for changes to take effect"
```



## Install kubectl

[Full Installation Guide][https://kubernetes.io/docs/tasks/tools/install-kubectl/]

Use the following script:

```sh
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
```

## Install minikube

[Full Installation Guide][https://minikube.sigs.k8s.io/docs/start/]

Use the following script:

```sh
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
```

## User specific steps

These needs to be done once.

1. Add your user in docker group

```
sudo usermod -aG docker "$USER"
```

**NOTE**: Logout/Login and verufy with `id` that your user is on group docker.

2. Add an empty kubernetes configuration

```
mkdir .kube
touch .kube/config
chmod 600 .kube/config
```

3. Start minikube

```
minikube start
```

4. Explore

