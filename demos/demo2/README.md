## Start minikube node

```sh
minikube start \
    --vm-driver=docker \
    --extra-config=kubeadm.pod-network-cidr=172.16.0.0/12 \
    --extra-config=kubelet.network-plugin=cni

```

## Start Andrea

```sh
kubectl apply -f https://github.com/vmware-tanzu/antrea/releases/download/v0.12.0/antrea.yml

```

## Create the containers 

```sh
# Alice
kubectl apply -f ./alice.yaml

# Malice
kubectl apply -f ./malice.yaml

# Snort
kubectl apply -f ./snort.yaml

```

## Enable port forwading to snort pod 

Find the name of the snort-pod

```sh
kubectl exec -n kube-system -it antrea-agent-s6gml -- ovs-vsctl show | grep snort 

```

Set up port mirroring to snort pod

```sh
kubectl exec -n kube-system -it antrea-agent-XXXX -- ovs-vsctl \
  -- --id=@p get port snort-XXXX \
  -- --id=@m create mirror name=m0 select-all=true output-port=@p \
  -- set bridge br-int mirrors=@m

```

## Inspect trafic inside bridge

Execute snort

```sh
kubectl exec -it snort -- snort -i eth0 -c /etc/snort/etc/snort.conf -A console

```

Ping google from malice 

```sh
kubectl exec -it malice -- ping -c3 8.8.8.8

```

Ping should appear on snort!

## Create snort pod with 2 network devices

We are going to download and install multus-cni on kubernetes installation

```sh
git clone https://github.com/intel/multus-cni.git && cd multus-cni
cat ./images/multus-daemonset.yml | kubectl apply -f -

```

Check that multus is live
```sh
kubectl get pods --all-namespaces | grep -i multus

```

Add a new interface for pods to use 
```sh
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
```sh
kubectl apply -f ./snort-multus.yaml

```

Expose antrea-controller service 
```sh
 kubectl expose deployment/antrea-controller --namespace=kube-system 
 kubectl get ep antrea-controller --namespace=kube-system


 ```

## Inspect traffic with script

Start snort on background
```sh
# -A unsock -l create a tmp/snort_alert sock
# Start snort on background
snort -i eth0 -A unsock -l /tmp -c /etc/snort/etc/snort.conf -D

```

Run script to connect to socket
```sh
python simple_icmp.py 


```
