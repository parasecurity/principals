# API functionality overview

The API service provides a simple way to deploy our primitives both externally and internally from the cluster. Each instruction follows the following structure `<action> <primitive> <arguments>`. In order to communicate with the centralize API service there is a client that we implimented. The client connects using __tls__ to the API and send the selected command. 

## Actions supported

`create`: Deploys the selected primitive with the arguments that are provided.

`delete`: Deletes the selected primitive.

## Primitives supported

`Canary`: The canary checks the response time from a specific heath point of a given website. When the response takes longer than a given threshold it automaticaly deploys a defence mechanism. 

`Canary-link`: Deploys a Daemoset that checks the link status of the a specific interface on the nodes. When the link becomes saturated it automaticaly deploys a defence mechanism.

`Detector`: Deploys a Deamonset that monitors the traffic inside the cluster through the mirrored port. When it detected a malicious spike in traffic to a specific IP address, it executes the user selected mitigation strategy.

`Detector-link`: Deploys a Deamonset that monitors the traffic inside the cluster through the mirrored port. When it detected a malicious spike in traffic, it executes the user selected mitigation strategy.

`Dga`: Delpoys a Daemoset that runs a service that uses a machine learning model to detect requests to domains created by domain generation algorithms. When a malicious request arrives from the bridge, through the mirrored port, it executes the user selected mitigation strategy.

`Snort`: Delpoys a Daemoset that spawns a snort pod. User selects the snort functionality and mitigation strategy.

`Honeypot`: Deploys a Daemoset that spawns a honeypot on each node. 

`Analyser`: Accepts as input a file containing the list of names we want to record requests to. Runs a service that uses scapy python module to analyse incoming traffic to the mirrored port. 

## Arguments Supported

`Block`: You block a given IP address.

`Unblock`: You unblock a given IP address.

`Throttle`: You rate-limit traffic of a given interface. The arguments supported are: 
  - port: the ovs port to be throttled
  - limit: the limit to be applied to the ovs port

`Forward`: You forward all outgoing  requests to a specific IP address. The supported arguments are:
  - original IP address
  - changed IP address
  - changed MAC address

`Tarpit`: You rate-limit traffic of a given interface. Queues need to be added to the port interfaces that you want to throttle traffic. After queues have been added, you provide the IP address you want to be tarpited. 

