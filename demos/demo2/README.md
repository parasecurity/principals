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

To build the images needed for the demo, run:


```
cd ..
./init.sh
cd demo2
```

To start the cluster, run:

```
cd ..
./deploy.sh
cd demo2
```

When all prerequisites are satisfied, you can start the demo with:

```sh
./run
```

## Pods created

`Agent server` daemonset runs a service that listens for ovs commands. When a command arrives it gets applied to the Open vSwitch bridge.

`Flow controller` daemonset runs a service that listens for requests from DGA detector. When a request arrives it is forwarded to Agent server.

`DGA detector` daemonset runs a service that uses a machine learning model to detect requests to domains created by domain generation algorithms. When a malicious request arrives from the bridge, through the mirrored port, the resorved IP address is forwarded to the Flow controller.

## Demo Main Contribution:
  - Flow controller daemonset
  - DGA daemonset 