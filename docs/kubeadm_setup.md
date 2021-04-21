# Install Kubernetes Cluster using kubeadm
Follow this documentation to set up a Kubernetes cluster on __Ubuntu 20.04 LTS__.

This documentation guides you in setting up a cluster with one master node and one worker node.

## Assumptions
|Role|FQDN|IP|OS|RAM|CPU|
|----|----|----|----|----|----|
|Master|kmaster.example.com|10.0.2.5|Ubuntu 20.04|2G|2|
|Worker|kworker.example.com|10.0.2.4|Ubuntu 20.04|1G|1|

## On both Kmaster and Kworker
Perform all the commands as root user unless otherwise specified
##### Disable Firewall
```sh
ufw disable
```
##### Disable swap
```sh
swapoff -a; sed -i '/swap/d' /etc/fstab
```
##### Update sysctl settings for Kubernetes networking
```sh
cat >>/etc/sysctl.d/kubernetes.conf<<EOF
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sysctl --system
```
##### Install docker engine
```sh
{
  apt install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
  add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
  apt update
  apt install -y docker-ce=5:19.03.10~3-0~ubuntu-focal containerd.io
}
```
### Kubernetes Setup
##### Add Apt repository
```sh
{
  curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
  echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list
}
```
##### Install Kubernetes components
```sh
apt update && apt install -y kubeadm=1.18.5-00 kubelet=1.18.5-00 kubectl=1.18.5-00
```
##### In case you are using LXC containers for Kubernetes nodes
Hack required to provision K8s v1.15+ in LXC containers
```sh
{
  mknod /dev/kmsg c 1 11
  echo '#!/bin/sh -e' >> /etc/rc.local
  echo 'mknod /dev/kmsg c 1 11' >> /etc/rc.local
  chmod +x /etc/rc.local
}
```

## On kmaster
##### Initialize Kubernetes Cluster
Update the below command with the ip address of kmaster
```sh
kubeadm init --apiserver-advertise-address=172.16.16.100 --pod-network-cidr=192.168.0.0/16  --ignore-preflight-errors=all
```

##### Cluster join command
```sh
kubeadm token create --print-join-command
```

##### To be able to run kubectl commands as non-root user
If you want to be able to run kubectl commands as non-root user, then as a non-root user perform these
```sh
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

## On Kworker
##### Join the cluster
Use the output from __kubeadm token create__ command in previous step from the master server and run here.

## Verifying the cluster
##### Get Nodes status
```sh
kubectl get nodes
```
##### Get component status
```sh
kubectl get cs
```

## On master
##### Deploy Antrea network
```sh
kubectl apply -f https://raw.githubusercontent.com/vmware-tanzu/antrea/main/build/yamls/antrea.yml
```

##### Get nodes status
```sh
kubectl get nodes
```

##### To provide access to k8-master to exec into pods
```sh
kubectl apply -f ./utils/role-deployment.yaml
```

## Usefull commands

To get pods from a specific node: just add `--field-selector spec.nodeName=k8-worker`

## Problem fixes

- Run join with sudo
- mv /etc/kubernetes/kubelet.conf /etc/kubernetes/admin.conf
- Add insecure registry to docker config
- Change ip to local regisrty
- Set multus master network to VMs master network

## Fix docker registry error

```sh
sudo cat >>/etc/docker/daemon.json<<EOF
{ "insecure-registries" : ["192.168.122.1:5000"] }
EOF
sudo systemctl restart docker.service
sudo systemctl restart docker.socket
```

## References
- Source Article: [Github link](https://github.com/justmeandopensource/kubernetes/blob/master/docs/install-cluster-ubuntu-20.md)
- Config not found: /etc/kubernetes/admin.conf error: [Stackoverflow](https://stackoverflow.com/questions/66213199/config-not-found-etc-kubernetes-admin-conf-after-setting-up-kubeadm-worker)