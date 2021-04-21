# Demo 5: Tarpitting incoming malicious connections

This demo shows that by using a machine learning model in conjunction with OpenFlows mirroring capabilities we can create a DGA detector that performs tarpitting on incoming malicious connections. 

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
- Tarpit: You rate-limit traffic of a given interface
  - ip: the ip to be throttled

## Tarpit command

In order to perform tarpitting on the connection from the malicious ip, we make extensive use of the queue capabilities of OVS. We limit the download speed of `wget maliciousIP/malware.exe` command by applying a max limit to the selected queue. The commands are:

- We create a queue on a selected interface with the name `100`:
```sh
ovs-vsctl set port <interface> qos=@newqos -- \
  --id=@newqos create qos type=linux-htb \
      queues:100=@queue -- \
  --id=@queue create queue other-config:max-rate=<max rate of 100 queue>
```

- Each time we want to apply a limit on an ip we need to pass the incoming and outgoing flow from the queue we created. We use the following ovs rules in order to achieve that:
```sh
ovs-ofctl add-flow br-int ip,nw_src=<malicious ip>,action=set_queue:100,goto_table:30
ovs-ofctl add-flow br-int ip,nw_dst=<malicious ip>,action=set_queue:100,goto_table:30
```

## Demo Main Contribution:
  - Added action Tarpit