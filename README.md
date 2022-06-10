# PRINCIPALS: PRogrammable INfrastructure for CounterIng Pugnacious Adversaries on a Large Scale

> Presenting PRINCIPALS a novel architecture for introducing safe programmability and adaptability in 5G networks, enabling more secure networks and endpoints relative to the current state of the art.

Quick Jump: [Installation](#installation)

## Installation

PRINCIPALS runs over kubernetes. It has multiple images that contain cybersecurity primives (DGA analysis, TLS fingerpinting etc) that are easily deployed and configured by a TAMELET handler.

### Kubernetes deployment

To deploy kubernetes you need to have some VMs or physical machines available. 
For the kubernete deployment we use ansible.

```Shell
git clone git@github.com:parasecurity/principals.git
cd principals
cd kubernetes-deployment
vim hosts.ini # Configure the IP addresses for master/workers nodes
vim group_vars/all.yml # Configure variables for deployment
vim group_vars/kube_cluster.yml # Configure the master node IP address
ansible-playbook start.yml
```

For a more detailed guide check out the following [README](https://github.com/parasecurity/principals/blob/main/cluster-deployment/README.md).


