# Add a new node 

Create the same username across all machines.
Install openssh-server on all nodes.

## Setup Passwordless SSH Login
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
On master node run:
```sh
docker run -d -p 5000:5000 --restart=always --name registry registry:2
```

## Enable insecure repository access
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
We first remove the tain from master and then add a tag
```sh
# Remove taint
kubectl taint node <master name> node-role.kubernetes.io/master:NoSchedule-
# Add label on master
kubectl label nodes <master name> dedicated=master
```

## Fix secondary interface problems
- https://stackoverflow.com/questions/52364785/kubernetes-service-external-ip-assigned-to-node-secondary-interface

## References
- https://www.digitalocean.com/community/tutorials/how-to-create-a-kubernetes-cluster-using-kubeadm-on-centos-7
- https://linuxize.com/post/how-to-add-and-delete-users-on-ubuntu-18-04/
- https://linuxize.com/post/how-to-enable-ssh-on-ubuntu-18-04/
- https://www.journaldev.com/30301/passwordless-ssh-login
