## Allow passwordless sudo
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
sudo   ALL=(ALL:ALL) ALL
# with
sudo  ALL=(ALL:ALL) NOPASSWD: ALL
```
Log-out and log in again to test the passwordless sudo

## References
- https://www.digitalocean.com/community/tutorials/how-to-create-a-kubernetes-cluster-using-kubeadm-on-centos-7
