# Demo 1: Block/Unblock traffic

This demo shows that by adding/removing OpenFlow rules we can control the
traffic between pods.

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
    argument: <ip address>,
    server_ip: <ip address of ovs-controller>
}
```

The possible actions are:

- Block: You block a given ip address 
- Unblock: You unblock a given ip address
