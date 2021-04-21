# Demo 4: Forward malicious traffic to honeypot

This demo shows that by using a machine learning model in conjunction with OpenFlows mirroring capabilities we can create a DGA detector that blocks all outgoing traffic on a malicious ip, except tcp connections to port 80 that are forwarded to honeypot pod.  

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


`Honeypot` daemonset runs a daemon that executes a low interaction honeypot. It is responsible to react to malicious tcp requests on port 80 from other pods on the cluster.

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
- Forward: You forward all outgoing tcp requests (on port 80)
  - ip: the ip whose traffic will be forwarded
  - honeypot_ip: the honeypot ip
  - honeypot_mac: the honeypot mac

## Forward command rules

For this demo we use the forward to honeypot capability. In order to a create this functionality we use three different OVS rules:

- Drop all packets to malicious ip.
```sh
ovs-ofctl add-flow br-int table=70,ip,nw_dst='<malicious ip>',priority=300,actions=drop
```

- Change all tcp requests (to port 80) on a malicious ip to honeypot ip/mac address. This rule is applied to table 70.
```sh
ovs-ofctl add-flow br-int table=70,tcp,tcp_dst=80,nw_dst='<malicious ip>',actions=mod_nw_dst:'<honeypot ip>',mod_dl_dst:'<honeypot mac>',goto_table:71
```

- Change honeypots response packet ip to original malicious ip address. This rule is applied to table 29.
```sh
ovs-ofctl add-flow br-int table=10,ip,dl_src='<honeypot mac>',nw_src='<honeypot ip>',actions=mod_nw_src:'<malicious ip>',goto_table:29
```

## Demo Main Contribution:
  - Added action Forward
  - Added honeypot deployment