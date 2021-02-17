## Create the containers 

```
# Alice
kubectl apply -f ./alice.yaml

# Malice
kubectl apply -f ./malice.yaml

# Snort
kubectl apply -f ./snort.yaml

```

## Enable port forwading to snort pod 

Find the name of the snort-pod

```
kubectl exec -n kube-system -it antrea-agent-s6gml -- ovs-vsctl show | grep snort 

```

Set up port mirroring to snort pod

```
kubectl exec -n kube-system -it antrea-agent-s6gml -- ovs-vsctl \
  -- --id=@p get port snort-XXXX \
  -- --id=@m create mirror name=m0 select-all=true output-port=@p \
  -- set bridge br-int mirrors=@m
```
## Inspect trafic inside bridge

Add ping alert in snort 

```
# Paste this in opt/rules 
alert icmp any any -> any any (msg:"Pinging...";sid:1000004;)

```
Execute snort

```
kubectl exec -it snort -- snort -i eth0 -c /etc/snort/etc/snort.conf -A console

```

Ping google from malice 

```
kubectl exec -it malice -- ping -c3 google.com

```

Ping and responce should appear on snort!

## Create snort pod with 2 network devices

We are going to download and install multus-cni on kubernetes installation

```
git clone https://github.com/intel/multus-cni.git && cd multus-cni
cat ./images/multus-daemonset.yml | kubectl apply -f -

```

Check that multus is live
```
kubectl get pods --all-namespaces | grep -i multus

```

Add an new interface for pods to use 
```
cat <\<EOF | kubectl create -f -
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: macvlan-conf
spec:
  config: '{
      "cniVersion": "0.3.0",
      "type": "macvlan",
      "master": "eth0",
      "mode": "bridge",
      "ipam": {
        "type": "host-local",
        "subnet": "192.168.1.0/24",
        "rangeStart": "192.168.1.200",
        "rangeEnd": "192.168.1.216",
        "routes": [
          { "dst": "0.0.0.0/0" }
        ],
        "gateway": "192.168.1.1"
      }
    }'
EOF

```

Create the new snort pod
```
kubectl apply -f ./snort-multus.yaml

```
