# Demo 8: Snort pod 

This demo shows that by using a snort pod in conjunction with OpenFlows mirroring capabilities we  create a logging service that logs all alerts created by traffic inside a Kubernetes node. We mirror all traffic that happens inside a Kubernetes node through the Open VSwitch port that connects to the snort pod. When the traffic triggers an alert, snort logs that alert to a `alert` file. The `alert` file is being broadcasted to a centralised syslog server pod and is being logged there.

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

`Snort` pod listens through the mirrored OVS port for traffic that triggers a snort alert. Runs as a daemon and each snort alert is written in the alert file. When the file is written it gets broadcasted through the secondary pod network to syslog server.

`Syslog server` pod runs a syslog server that listens for logs to be recorded. When a snort alert is broadcasted through the secondary network to the server, it is recorded on a log file.
