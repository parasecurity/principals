# Demo 2 - DGA detection capability

This demo shows that by using a machine learning model in conjunction with OpenFlows mirroring capabilities we can create a DGA detector on kubernetes.  

We assume that the following software is already installed:

- docker
- kubectl
- minikube

We also assume that the user is on docker group. If not run:

```sh
sudo usermod -aG docker "$USER"
```

**NOTE**: Logout/Login and verify with `id` that your user is on group docker.

Also you should at least have on your home directory an empty kubernetes
configuration:

```sh
mkdir .kube
touch .kube/config
chmod 600 .kube/config
```

When all prerequisites are satisfied, you can start the demo with:

```sh
./run
```

## Pods created

`DGA detector` pod runs a service that uses a machine learning model to detects requests to domains created by domain generation algorithms. When a malicious request arrives from the bridge, through the mirrored port, the resorved IP address is forwarded to the Flow controller.

`Flow controller` pod runs a service that listens for  requests from DGA detector. When a request arrives it is forwared to OVS controller.

`OVS controller` pod runs a service on the same pod with the Open vSwitch bridge. It listens for commands from Flow controller. When a command arrives it gets applied to the Open vSwitch bridge.
