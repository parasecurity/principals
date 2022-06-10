# Kubernetes Cluster automation

The purpose of this repo is to automate the creation of a Kubernetes cluster
for RnD.

It uses VMs as cluster nodes. VM management is handled using _vagrant_ and the
provisioning using _ansible_.

Currently, the base system for all nodes is:

- ubuntu 20.04
- 8GB ram
- 8 cores
- user = vagrant
- user has passwordless sudo

Provisioning adds the following:

- git
- tig
- vim
- htop
- jq

Bring VMs up:

```
vagrant up
```

Halt VMs:

```
vagrant halt
```

Connect to a VM:

```
vagrant ssh master
vagrant ssh worker0
```

Full documentation can be found [here](https://www.vagrantup.com/docs/index)

Please check the provided Vagrantfile for further technical details/choices.

**NOTE**: The actual docker/kubernetes installation is not included yet! But it
should be trivial to set it up after the VMs are up.

**NOTE**: VMs end up all with the same IP on eth0. Vagrant is not responsible
for VM ip assignments. This is actually provider specific. Since we use
VirtualBox as a provider, _vboxmanage_ must be used to crrectly manage IP
assignments and this is out of scope for now.

See [also](https://www.vagrantup.com/docs/providers/virtualbox/configuration#vboxmanage-customizations)

In order to work around this for setting up a cluster, an additional private
network is added with manual IP assignment so each node gets a unique IP.
