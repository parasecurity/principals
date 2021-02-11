# Open vSwitch on Ubuntu 18.04

## Install Open vSwitch:

```
sudo apt install \
    bridge-utils \
    openvswitch-common \
    openvswitch-switch
```

## First example using only network namespaces

Based on info from:

EdX
LinuxFoundationX: LFS165x
Introduction to Open Source Networking Technologies 

End result:

+---------------------------------------------------------------+
|                                                               |
|                    +-----------------------+                  |
|              veth1 |                       | veth3            |
|                    | Open vSwitch: br0     |                  |
|          +---------+ 10.10.10.3/24         +--------+         |
|          |         |                       |        |         |
|    veth0 |         +-----------------------+        | veth2   |
|          |                                          |         |
|    +-----+---------------+         +----------------+----+    |
|    |                     |         |                     |    |
|    |  vrf0               |         |  vrf1               |    |
|    |  10.10.10.1/24      |         |  10.10.10.2/24      |    |
|    |                     |         |                     |    |
|    +---------------------+         +---------------------+    |
|                                                               |
|                                                               |
|    Host: Ubuntu 18.04                                         |
|                                                               |
+---------------------------------------------------------------+

Add two network namespaces:

```
sudo ip netns add vrf0
sudo ip netns add vrf1
ip netns
```

Add two veth pairs:

```
sudo ip link add veth0 type veth peer name veth1
sudo ip link add veth2 type veth peer name veth3
ip link
```

Move (isolate) two veths to separate network namespaces:

```
sudo ip link set veth0 netns vrf0
sudo ip link set veth2 netns vrf1
sudo ip netns exec vrf0 ip link
sudo ip netns exec vrf1 ip link
ip link
```

Assign IP address at the isolated veths and bring them up:

```
sudo ip netns exec vrf0 ip address add 10.10.10.1/24 dev veth0
sudo ip netns exec vrf0 ip link set veth0 up
sudo ip netns exec vrf0 ip address

sudo ip netns exec vrf1 ip address add 10.10.10.2/24 dev veth2
sudo ip netns exec vrf1 ip link set veth2 up
sudo ip netns exec vrf1 ip address
```

No connection yet between them, they are isolated:

```
ping -c 3 -W 1 10.10.10.1
ping -c 3 -W 1 10.10.10.2
sudo ip netns exec vrf0 ping -c 3 -W 1 10.10.10.2
sudo ip netns exec vrf1 ping -c 3 -W 1 10.10.10.1
```

Check current routes for all network namespaces:

```
ip route
sudo ip netns exec vrf0 ip route
sudo ip netns exec vrf1 ip route
```

Create an OVS bridge:

```
sudo ovs-vsctl add-br br0
sudo ovs-vsctl show
```

Connect the remaining veths to the bridge:

```
sudo ovs-vsctl add-port br0 veth1
sudo ovs-vsctl add-port br0 veth3
sudo ovs-vsctl show
```

Bring everything together:

```
sudo ip address add 10.10.10.3/24 dev br0
sudo ip link set br0 up
sudo ip link set veth1 up
sudo ip link set veth3 up
```

Verify the L2 switch operation:

```
ping -c 3 -W 1 10.10.10.1
ping -c 3 -W 1 10.10.10.2
sudo ip netns exec vrf0 ping -c 3 -W 1 10.10.10.2
sudo ip netns exec vrf1 ping -c 3 -W 1 10.10.10.1
```

Forwarding table:

```
sudo ovs-appctl fdb/show br0
```

Clean up:

```
sudo ip link delete veth1
sudo ip link delete veth3
sudo ip netns exec vrf0 ip link delete veth0
sudo ip netns exec vrf1 ip link delete veth2
sudo ip netns del vrf0
sudo ip netns del vrf1
sudo ovs-vsctl del-br br0
```

## Second example using containers

Based on:

- first example
- https://github.com/sarun87/examples/blob/master/networking/docker_point2point.sh
- https://archive.nanog.org/sites/default/files/mon.tutorial.wallace.openflow.31.pdf

End result:

+---------------------------------------------------------------+
|                                                               |
|                    +-----------------------+                  |
|              veth1 |                       | veth3            |
|                    | Open vSwitch: br0     |                  |
|          +---------+ 10.10.10.3/24         +--------+         |
|          |         |                       |        |         |
|    veth0 |         +-----------------------+        | veth2   |
|          |                                          |         |
|    +-----+---------------+         +----------------+----+    |
|    |                     |         |                     |    |
|    |  docker: guest0     |         |  docker: guest1     |    |
|    |  10.10.10.1/24      |         |  10.10.10.2/24      |    |
|    |  busybox            |         |  busybox            |    |
|    +---------------------+         +---------------------+    |
|                                                               |
|                                                               |
|    Host: Ubuntu 18.04                                         |
|                                                               |
+---------------------------------------------------------------+


Dependencies:

- docker
- network to pull busybox if not already available

Create two minimal guests without network:

```
docker run --name guest0 --network none -td busybox
docker run --name guest1 --network none -td busybox
```

Setup network namespace for each container:

```
sudo mkdir -p /var/run/netns

g0_pid="$(docker inspect --format '{{.State.Pid}}' guest0)"
g1_pid="$(docker inspect --format '{{.State.Pid}}' guest1)"

sudo ln -s /proc/$g0_pid/ns/net /var/run/netns/guest0
sudo ln -s /proc/$g1_pid/ns/net /var/run/netns/guest1
```

Add two veth pairs:

```
sudo ip link add veth0 type veth peer name veth1
sudo ip link add veth2 type veth peer name veth3
```

Move one veth to each container network namespace:

```
sudo ip link set veth0 netns guest0
sudo ip link set veth2 netns guest1
```

Assign IP address at the container veths and bring them up:

```
sudo ip netns exec guest0 ip address add 10.10.10.1/24 dev veth0
sudo ip netns exec guest0 ip link set veth0 up

sudo ip netns exec guest1 ip address add 10.10.10.2/24 dev veth2
sudo ip netns exec guest1 ip link set veth2 up
```

Create a bridge at the host and connect the remaining veths:

```
sudo ovs-vsctl add-br br0
sudo ovs-vsctl add-port br0 veth1
sudo ovs-vsctl add-port br0 veth3
sudo ovs-vsctl show
```

Verify there is no communication yet:

```
ping -c3 -W 1 10.10.10.1
ping -c3 -W 1 10.10.10.2
docker exec guest0 ping -c3 -W 1 10.10.10.2
docker exec guest1 ping -c3 -W 1 10.10.10.1
```

Bring everything up:

```
sudo ip address add 10.10.10.3/24 dev br0
sudo ip link set br0 up
sudo ip link set veth1 up
sudo ip link set veth3 up
```

Communication works for everyone:

```
ping -c3 10.10.10.1
ping -c3 10.10.10.2
docker exec guest0 ping -c3 10.10.10.2
docker exec guest1 ping -c3 10.10.10.1
```

Forwarding table:

```
sudo ovs-appctl fdb/show br0
```

Show the current flows:

```
sudo ovs-ofctl dump-flows br0
```

Remove them and verify that no connection is working:

```
sudo ovs-ofctl del-flows br0
sudo ovs-ofctl dump-flows br0
ping -c3 -W1 10.10.10.1
ping -c3 -W1 10.10.10.2
docker exec guest0 ping -c3 -W1 10.10.10.2
docker exec guest1 ping -c3 -W1 10.10.10.1
```

Add two new flows and verify that only guest0 <-> guest1 can talk:

```
sudo ovs-ofctl add-flow br0 idle_timeout=180,priority=10,in_port=1,actions=output:2
sudo ovs-ofctl add-flow br0 idle_timeout=180,priority=10,in_port=2,actions=output:1
ping -c3 -W1 10.10.10.1
ping -c3 -W1 10.10.10.2
docker exec guest0 ping -c3 10.10.10.2
docker exec guest1 ping -c3 10.10.10.1
```

Clean up:

```
docker kill guest0
docker kill guest1
docker rm guest0
docker rm guest1
sudo unlink /var/run/netns/guest0
sudo unlink /var/run/netns/guest1
sudo ovs-vsctl del-br br0
```

