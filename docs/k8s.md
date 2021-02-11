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

At this point all tools are setup. The first step is to start minikube:

```
minikube start
```

And then experiment with it. To stop it:

```
minikube stop
```

If clean up is needed:

```
minikude delete
```

## Using Antrea CNI

Start a clean minikube with this options:

```
minikube start \
    --vm-driver=docker \
    --network-plugin=cni \
    --extra-config=kubeadm.pod-network-cidr=172.16.0.0/12 \
    --extra-config=kubelet.network-plugin=cni
```

Install antrea:

```
kubectl apply \
    -f https://github.com/vmware-tanzu/antrea/releases/download/v0.12.0/antrea.yml
```

Display info:

```
watch kubectl get -A po
```

Wait until antrea status is 'Running'.
When it is running, exit watch and run the following script:

```
#!/usr/bin/env bash

readonly ANTREA_AGENT=$(kubectl get -A po | grep "antrea-agent" | awk '{print $2}')

echo "alias antrea=\"kubectl -n kube-system exec -it $ANTREA_AGENT -c antrea-agent -- \"" | tee -a $HOME/.bash_aliases
```

And then:

```
source $HOME/.bash_aliases
```

At this point antrea is setup and you can bring a shell inside the antrea
agent container using:

```
antrea bash
```

Or run commands, e.g.:

```
antrea ovs-vsctl show
antrea ovs-ofctl dump-flows br-int
```

## Demo 1 - Block/Unblock traffic

Create a demo directory:

```
mkdir -p $HOME/demo1
```

And create inside the following two files:

alice.yaml:

```
apiVersion: v1
kind: Pod
metadata:
  name: alice
spec:
  containers:
  - name: alice
    image: praqma/network-multitool
    command:
      - sleep
      - infinity
    imagePullPolicy: IfNotPresent
```

malice.yaml:

```
apiVersion: v1
kind: Pod
metadata:
  name: malice
spec:
  containers:
  - name: malice
    image: praqma/network-multitool
    command:
      - sleep
      - infinity
    imagePullPolicy: IfNotPresent
```

Assuming antrea is up and running, create the demo pods:

```
kubectl create -f demo1/alice.yaml
kubectl create -f demo1/malice.yaml
```

Get the pods IP's:

```
kubectl exec -it alice -- ip a | grep "inet.*eth0" | awk '{print $2}'
kubectl exec -it malice -- ip a | grep "inet.*eth0" | awk '{print $2}'
```

Ping alice from malice using the correct ip:

```
kubectl exec -it malice -- ping -c3 172.16.0.xxx
```

Find from the antrea ovs switch the port malice is using:

```
antrea ovs-vsctl show | grep "Port malice" | awk '{print $2}'
```

Now get all the current flow rules for malice:

```
antrea ovs-ofctl dump-flows br-int | grep malice
```

Add a new rule to drop malice packets:

```
antrea ovs-ofctl add-flow br-int in_port="malice-xxxxxx",nw_src=172.16.0.xxx,actions=drop
```

Verify it is added:

```
antrea ovs-ofctl dump-flows br-int | grep malice
```

Ping alice from malice to verify that is does not work:

```
kubectl exec -it malice -- ping -c3 -W1 172.16.0.xxx
```

Remove the drop rule:

```
antrea ovs-ofctl del-flows --strict br-int in_port="malice-xxxxxx",nw_src=172.16.0.xxx
```

Verify it is removed:

```
antrea ovs-ofctl dump-flows br-int | grep malice
```

Ping alice from malice to verify it works again:

```
kubectl exec -it malice -- ping -c3 172.16.0.xxx
```

