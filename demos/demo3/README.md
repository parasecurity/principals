# Demo 3: Throttle traffic

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

## Pods created

`Agent server` daemonset runs a service that listens for ovs commands. When a command arrives it gets applied to the Open vSwitch bridge.

`Flow controller` pod provides an interface to send commands to Agent server.

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
  - ip: the ip to be blocked
- Unblock: You unblock a given ip address
  - ip: the ip to be unblocked
- Throttle: You rate-limit traffic of a given interface
  - port: the ovs port to be throttled
  - limit: the limit to be applied to the ovs port

## Demo Main Contribution:
 - Added action Throttle
