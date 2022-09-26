# Analyzer Primitive
> This is the Analyzer Primitive created for the PRINCIPALS project. In this Primitive we want to monitor specific domains, analyze traffic from/to them and store specific information about that traffic.


## Mirroring

This Primitive uses a mirroring of the whole nodes' networking traffic. For the mirroring to be realized you need to create a docker container. This container will use the image of Antrea as a base image because we want to have a Volume-mount with Antrea-Agent inside the K8s Cluster. Inside this file will be placed the script used for the appliance of the mirroring command to the bridge. This whole implementation can be found inside the mirroring directory(The Dockerfile needed for the container and the script needed for the mirroring).



## Installation
In order to install this Primitive inside your own Kubernetes Cluster, you should first create a Docker image that has all the necessary packages installed inside it. This can be done via the Dockerfile, using inside the directory that rests your Dockerfile alongside the necessary files, the command:

```Shell
docker build -t analyzer .
```
Inside this container lies the monitor.py script that provides the necessary functionality for this primitive.

In order to deploy it to your own Kubernetes Cluster, you need to apply the yaml file. The yaml file contains an Init Container component for the mirroring to execute before the Analyzer Pod is deployed. Before applying the yaml file provided, you should first change the image variable inside it to your own image's name(both the mirroring image and the Analyzer image). The command needed to apply the yaml file is:

```Shell
kubectl apply -f <yaml file>
```
where "yaml file" is the name of your yaml file. When deployed the Pod will have the whole nodes' traffic mirrored to its eth0 interface.

## Implementation

The necessary functionality for this Primitive is provided by the monitor.py script. With the help of Scapy, a sniffing process begins with a filter of port 53, since we’re only interested in DNS Requests. When a DNS request matches with a domain from the list of domains we get from the input file(domains.txt in our demo), the script creates the designated object with the information. When this job is done, through a TCP
socket, the pod will communicate with the Flow-Server pod, or any other logger pod that we desire, and will exchange these information.

It’s important to note, that since our first interface is used for the mirroring, the communication between our pod and the Flow-Server pod, is carried out through the secondary interface of those two pods. The secondary interface is
created with the use of Multus inside the yaml file.
