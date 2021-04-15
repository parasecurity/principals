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

To build the images needed for the demo, run:

```
cd ..
./init.sh
cd demo1
```

To start the cluster, run:

```
cd ..
./deploy.sh
cd demo1
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
    argument: <json>
}
```

The argument is a json with all the arguments needed by the action

The possible actions with their arguments are:

- Block: You block a given ip address
  - malicious_ip: the ip to be blocked
- Unblock: You unblock a given ip address
  - ip: the ip to be blocked


Demo Main Contribution: Agent server daemonset that gives us the ability to apply our own flows to the system