## Create VMs

Use this [repo](git@kition.mhl.tuc.gr:tsi-group/kubecluster.git) in order to generate the new VMs.
There is a nessasary port configuration on the _VagrantFile_ and  _worker.vm.network/master.vm.network_
in order to check there is no overlapping network.

## Versions 
Tested on ansible version 2.9.27.
Steps to install [ansible](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html)

## Configure ansible deployment

On the group_var folder set the registry ip, pod_network_cidr, and master_ip on the correct values.

## Deploy ansible

In order to deploy the ansible configuration and create the cluster:
```sh
ansible-playbook start.yml
```

## Setup Passwordless SSH Login

### If you running a vagrant deployment

first find the forwared ports 
```sh
vagrant ssh-config
```

Write down the ports and then run
```
ssh-copy-id vagrant@localhost <port>
# For password use vagrant
```

### For any other VM deployment
ssh to ansible control machine
```sh
ssh-copy-id remote_username@server_ip_address
```

You will be asked if you are sure you want to continue connecting. Type yes and hit ‘Enter’
```sh
The authenticity of host '173.82.2.236 (173.82.2.236)' can't be established.
ECDSA key fingerprint is SHA256:U4aOk0p30sFjv1rzgh73uhGilwJ2xtG205QFqzB9sns.
Are you sure you want to continue connecting (yes/no)? yes
```

Next, you will be prompted for the remote system’s password. Type the password and hit ‘Enter’
```sh
root@173.82.2.236's password:

Number of key(s) added: 1

Now try logging into the machine, with:   "ssh 'root@173.82.2.236'"
and check to make sure that only the key(s) you wanted were added.
```

## Setup passwordless sudo
NOTE: if you used kubecluster's vagrant, then it is already set up

ssh to the remote user
```sh
ssh <user>@<ip address>
```

On the remote server:
```sh
sudo visudo
```

Edit the line:
```sh
%sudo  ALL=(ALL:ALL) ALL
# with
%sudo  ALL=(ALL:ALL) NOPASSWD: ALL
```
Log-out and log in again to test the passwordless sudo

## Start local registry on master node
NOTE: it is automated by this playbook

On master node run:
```sh
docker run -d -p 5000:5000 --restart=always --name registry registry:2
```

## Enable insecure repository access
NOTE: it is automated by this playbook

On all nodes on the cluster
```sh
# Change the ip with the ip of the registry
sudo cat >>/etc/docker/daemon.json<<EOF
{ "insecure-registries" : ["192.168.122.1:5000"] }
EOF
sudo systemctl restart docker.service
sudo systemctl restart docker.socket
```

## Add autocompletion on kubectl
Just run on the node you want to enable kubectl autocompletion
```sh
sudo apt install bash-completion
echo 'source <(kubectl completion bash)' >>~/.bashrc
kubectl completion bash >/etc/bash_completion.d/kubectl
```

## Enable node deployment on master node
NOTE: it is automated by this playbook

We first remove the tain from master and then add a tag
```sh
# Remove taint
kubectl taint node <master name> node-role.kubernetes.io/master:NoSchedule-
# Add label on master
kubectl label nodes <master name> dedicated=master
```
## deploy NetworkAttachmentDefinition
You need to deploy multus daemonset before running any demos
Ansible will install https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/deployments/multus-daemonset.yml
which is tested. You can remove it and deploy the yaml of principals repo ( which is an older version )

```sh

kubectl delete -f https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/deployments/multus-daemonset.yml
# wait until it is removed
kubectl apply -f /path/to/principals/demos/demo1/yaml/multus-daemonset.yml
```
## When deploying localhost

At the beginning modify the hosts.ini file:
```sh
[master]
127.0.0.1 ansible_connection=local

[kube_cluster:children]
master
```
And after the installation:
```sh
vim /etc/default/kubelet
# Add the correct node-ip on the extra arguments
KUBELET_EXTRA_ARGS="--node-ip=10.9.9.2"
```

## Ansible gathering facts - very slow
```
# Add to the ansible.cfg
gather_subset=!hardware
```

## TODO
- Make deployment install kubernetes version 1.23.1 and not the latest. Kubernetes drops docker support on version 1.24.

## References
- https://www.digitalocean.com/community/tutorials/how-to-create-a-kubernetes-cluster-using-kubeadm-on-centos-7
- https://linuxize.com/post/how-to-add-and-delete-users-on-ubuntu-18-04/
- https://linuxize.com/post/how-to-enable-ssh-on-ubuntu-18-04/
- https://www.journaldev.com/30301/passwordless-ssh-login
