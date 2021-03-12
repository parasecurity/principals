# Demo 1: Throttle traffic

This demo shows that by adding OpenFlow rules to an interface we can rate-limit traffic by a specific pod. There are two values to set:

- ingress\_policing\_rate: the maximum rate (in Kbps) that this pod should be allowed to send
- ingress\_policing\_burst: a parameter to the policing algorithm to indicate the maximum amount of data (in Kb) that this interface can send beyond the policing rate. 

We assume that the following software is already installed:

- docker
- kubectl
- minikube

We also assume that the user is on docker group. If not run:

```
sudo usermod -aG docker "$USER"
```

**NOTE**: Logout/Login and verify with `id` that your user is on group docker.

Also you should at least have on your home directory an empty kubernetes
configuration:

```
mkdir .kube
touch .kube/config
chmod 600 .kube/config
```

When all prerequisites are satisfied, you can start the demo with:

```
./run
```

## Current interface

Flow controller supports input through a `.json` file.

The json fields are:
```
json = {
    action: <action>,
    argument: <ip address/interface>,
    server_ip: <ip address of ovs-controller>
}
```

The possible actions are:

- Block: You block a given ip address 
- Unblock: You unblock a given ip address
- Throttle: You rate-limit traffic of a given interface. Needs two arguments, the selected `port` and the speed `limit` to be applied.
