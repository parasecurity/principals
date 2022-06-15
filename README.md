# PRINCIPALS: PRogrammable INfrastructure for CounterIng Pugnacious Adversaries on a Large Scale

> Presenting PRINCIPALS a novel architecture for introducing safe programmability and adaptability in 5G networks, enabling more secure networks and endpoints relative to the current state of the art.

Quick Jump: [Installation](#installation) | [Primitives](#primitives) | [Demos](#demos)

## Installation

PRINCIPALS runs over kubernetes. It has multiple images that contain cybersecurity primives (DGA analysis, TLS fingerpinting etc) that are easily deployed and configured by a TAMELET handler.

### Kubernetes deployment

To deploy kubernetes you need to have some VMs or physical machines available. 
If you do not have some running VMs or physical machines check out the following guide [README](https://github.com/parasecurity/principals/blob/main/cluster-deployment/vm-deployment/README.md) in the [vm-deployment](https://github.com/parasecurity/principals/tree/main/cluster-deployment/vm-deployment) folder.

For the kubernetes deployment we use ansible.

```Shell
git clone git@github.com:parasecurity/principals.git
cd principals
cd cluster-deployment/kubernetes-deployment
vim hosts.ini # Configure the IP addresses for master/workers nodes
vim group_vars/all.yml # Configure variables for deployment
vim group_vars/kube_cluster.yml # Configure the master node IP address
ansible-playbook start.yml
```

For a more detailed guide check out the following [README](https://github.com/parasecurity/principals/blob/main/cluster-deployment/kubernetes-deployment/README.md).

## Primitives

It contains a list of different security primitives packaged in a docker container, bundled with yamls to deploy on a 5G installation quickly.

### Domain Analyser

The domain [analyzer](https://github.com/parasecurity/principals/tree/main/images/analyser) primitive contains a python script that analyses incoming traffic. The analyzer checks against a list a provider list of domains, and when a domain match occurs, it forwards it to a third server.
 

## Demos
