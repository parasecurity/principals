# Demo 6: Flow recording 

This demo shows that by using OpenFlows mirroring capabilities we can create a flow analyser that forwards packet info to a pod that records them. In order for the packet info to be forwarded, the domain of the packet should appear in a predefined list provided by admin. Users are able to access the logs from _outside_ the cluster with a client.  

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

`Analyser` pod accepts as input a file containing the list of names we want to record requests to. Runs a service that uses scapy python module to analyse incoming traffic to the mirrored port. When the domain of the packet exists in the predefined list, specific info about the packet gets forwarded to the `logger` pod. 

`Logger` pod runs a service that listens for info packaged as a json from the `analyser` pod. When a json arrives from the analyser the data are recorded to a `data.log` file in order to accessible by admin in the future. When admin wants to access the logs in the future he can access them both inside and outside of the cluster with the provided client app.
