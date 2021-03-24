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

`DGA detector` pod runs a service that uses a machine learning model to detects requests to domains created by domain generation algorithms. When a malicious request arrives from the bridge, through the mirrored port, the resolved IP address is forwarded to the Flow controller.

`Flow controller` pod runs a service that listens for  requests  from a DGA detector. When a request arrives it is forwarded to the OVS controller.

`OVS controller` pod runs a service on the same pod with the Open vSwitch bridge. It listens for commands from the Flow controller. When a command arrives it gets applied to the Open vSwitch bridge.

`Honeypot` pod runs a daemon that executes a low interaction honeypot. It is responsible to react to malicious tcp requests on port 80 from other pods on the cluster.

## Possible commands 

- Block \<ip\>: Blocks all traffic with a specific ip address on the cluster.
- Unblock \<ip\>: Unblock all traffic with a specific ip address on the cluster.
- Throttle \<port\> \<limit\>: Throttles traffic to a specific limit on a port on the ovs-bridge.
- Forward \<ip\> \<honeypot ip\> \<honeypot mac\>: Forwards all outgoing tcp requests (on port 80) on a malicious ip to honeypot pod. All other requests to malicious ip are dropped.

## Forward command rules

For this demo we use the forward to honeypot capability. In order to a create this functionality we use three different OVS rules:

- Drop all packets to malicious ip.
```sh
os.system('ovs-ofctl add-flow br-int table=70,ip,nw_dst='<malicious ip>',priority=300,actions=drop')
```

- Change all tcp requests (to port 80) on a malicious ip to honeypot ip/mac address. This rule is applied to table 70.
```sh
os.system('ovs-ofctl add-flow br-int table=70,tcp,tcp_dst=80,nw_dst='<malicious ip>',actions=mod_nw_dst:'<honeypot ip>',mod_dl_dst:'<honeypot mac>',goto_table:71')
```

- Change honeypots response packet ip to original malicious ip address. This rule is applied to table 29.
```sh
os.system('ovs-ofctl add-flow br-int table=10,ip,dl_src='<honeypot mac>',nw_src='<honeypot ip>',actions=mod_nw_src:'<malicious ip>',goto_table:29')
```

For more info about OVS tables existing on the current antrea configuration [visit](https://github.com/vmware-tanzu/antrea/blob/main/docs/design/ovs-pipeline.md).