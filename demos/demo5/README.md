# Demo 5

This demo shows that by using a machine learning model in conjunction with OpenFlows mirroring capabilities 

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

Before each run of the demo you must clean the previous minikube configuration and delete previous created DGA, flow control images. To do that, run:

```sh
minikube stop
minikube delete
docker images 
# Find tsi-flow-control, tsi-dga images id
docker rmi -f \<tsi-flow-control image id\>, \<tsi-dga image id\>
```

When all prerequisites are satisfied, you can start the demo with:

```sh
./run
```

## Pods created

`DGA detector` pod runs a service that uses a machine learning model to detects requests to domains created by domain generation algorithms. When a malicious request arrives from the bridge, through the mirrored port, the  reserved IP address is forwarded to the Flow controller.

`Flow controller` pod runs a service that listens for  requests  from a DGA detector. When a request arrives it is forwarded to the OVS controller.

`OVS controller` pod runs a service on the same pod with the Open vSwitch bridge. It listens for commands from  from the Flow controller. When a command arrives it gets applied to the Open vSwitch bridge.

## Possible commands 

- Block \<ip\>: Blocks all traffic with a specific ip address on the cluster.
- Unblock \<ip\>: Unblock all traffic with a specific ip address on the cluster.
- Throttle \<port\> \<limit\>: Throttles traffic to a specific limit on a port on the ovs-bridge.
- Forward \<ip\> \<honeypot ip\> \<honeypot mac\>: Forwards all outgoing tcp requests (on port 80) on a malicious ip to honeypot pod. All other requests to malicious ip are dropped.
- TODO: add turping command
