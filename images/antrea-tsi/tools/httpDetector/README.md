# DDoS Detector Primitive
> This is the DDoS Detector Primitive created for the PRINCIPALS project. A Distributed Denial of Service (DDoS) is a kind of attack that plans to overflood the traffic of server, service or network. The way to achieve such a malicious
act is by overwhelming the target or its surrounding infrastructure with a flood of
Internet traffic. For the purpose of defending such an attack, this Primitive was created.



## Mirroring

This Primitive uses a mirroring of the whole nodes' networking traffic. For the mirroring to be realized you need to create a docker container. This container will use the image of Antrea as a base image because we want to have a Volume-mount with Antrea-Agent inside the K8s Cluster. Inside this file will be placed the script used for the appliance of the mirroring command to the bridge. This whole implementation can be found inside the mirroring directory(The Dockerfile needed for the container and the script needed for the mirroring).



## Installation
In order to install this Primitive inside your own Kubernetes Cluster, you should first create a Docker image that has all the necessary packages installed inside it. This can be done via the Dockerfile, using inside the directory that rests your Dockerfile alongside the necessary files, the command:

```Shell
docker build -t ddosdetector .
```
Inside this container lies the httpDetector.go script that provides the necessary functionality for this primitive.

In order to deploy it to your own Kubernetes Cluster, you need to apply the yaml file. The yaml file contains an Init Container component for the mirroring to execute before the ddosdetector is deployed. Before applying the yaml file provided, you should first change the image variable inside it to your own image's name(both the mirroring image and the ddosdetector image). The command needed to apply the yaml file is:

```Shell
kubectl apply -f <yaml file>
```
where "yaml file" is the name of your yaml file. When deployed the Pod will have the whole nodes' traffic mirrored to its eth0 interface.


## Implementation

The necessary functionality for this Primitive is provided by the httpDetector.go script. Using the pcap library, we create a monitoring interface that starts with a BPF filter of net 127.0.0.1, in other words that monitors the loopback. At the
same time a TCP socket is open for listening incoming traffic. This TCP socket and any other incoming or outgoing communication is happening through the secondary interface. The secondary interface is created with the use of Multus inside
the yaml file. Whenever the Canary, a pod that already exists in the PRINCIPALS project and won’t be explained thoroughly, detects any weird traffic, it will send a message to our pod, through the TCP socket, with the IP that creates it.
In every message our listener gets, each IP is extracted and added to the array of IPs that are being monitored. Bear in mind that all of this is happening with the necessary use of Mutexes(locks and unlocks) since Go is a multi-threaded
language, as already mentioned. In every new IP we get, the BPF filter changes
to the appropriate value in contemplation of monitoring this IP as well. Simultaneously, two timers have been set, to check periodically the incoming traffic of the IPs that we monitor. Through a variety of ifs, each packet is adding value to
a number of counters (tcpCounter, udpCounter, HttpCounter, etc), depending on
its characteristics. At the end of the first timer if the percentage of traffic is unusually high for a certain type, then we know we are under a
certain type of attack. In that case, our pod will have to inform the Flow-Server about the attack that we are under. So the pod will send to the
Flow-server through a TCP socket the IP of the malicious pod and the command to be executed by the bridge. If no attack is found by the time the second timer strikes, then the IP is cleared from suspision(clear from the monitoredIPs array) and the BPF filter changes appropriately.

It’s important to note, that since our first interface is used for the mirroring, the communication between our pod and the Flow-Server pod, is carried out through the secondary interface of those two pods. The secondary interface is
created with the use of Multus inside the yaml file.
