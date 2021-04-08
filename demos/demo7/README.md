# Demo 7: Multi Node scalable DGA detection 

This demo shows that by using a machine learning model in conjunction with OpenFlows mirroring capabilities we can create a DGA detector that blocks all outgoing traffic to a malicious IP address from all nodes inside a cluster. When a DGA detector registers a malicious domain it forwards the resolved IP address to the flow client through a TCP connection. The flow client packages the IP address into a JSON file (that contains block command and resolved IP address) to the flow server through a TCP connection. The flow server forwards the JSON file to all nodes in the cluster and then gets applied through the services running inside antrea agents pods.

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

`DGA detector` pod runs a service that uses a machine learning model to detect requests to domains created by domain generation algorithms. When a malicious request arrives from the bridge, through the mirrored OVS port, the resolved IP address is forwarded to the Flow client through an established TCP connection.

`Flow client` pod runs a python script that listen for incoming IP addresses from DGA detector pod and the packages them into a JSON file of the following format:
```python
obj = {
        "action": block,
        "argument": <malicious IP address>,
    }
```
Then the package gets forwarded through an established TCP connection to the flow server that is running on the master cluster.

`Flow server` pod runs a golang program that listens for senders on TCP port (12345) and for receivers on TCP port (23456). When a sender sends a message, that message is forwarded to all connected receivers.

`OVS controller` runs a service on the same pod with the Open vSwitch bridge. It listens for commands from the Flow server. When the block command arrives it gets applied to the OVS bridge.
