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

Delete the previous config

```sh
minikube delete

```

Start minikube node

```sh
minikube start \
    --vm-driver=docker \
    --extra-config=kubeadm.pod-network-cidr=172.16.0.0/12 \
    --extra-config=kubelet.network-plugin=cni

```

Start Andrea

```sh
kubectl apply -f https://github.com/vmware-tanzu/antrea/releases/download/v0.12.0/antrea.yml

```

Download and install multus-cni 

```sh
kubectl apply -f https://raw.githubusercontent.com/intel/multus-cni/master/images/multus-daemonset.yml

```

Check that pod are live 

```sh
watch kubectl get pods --all-namespaces 

```

Create network interface

```sh 
kubectl create -f ./2port.yaml

```

Create a snort and alice pod with 2 port interfaces

```sh
# Snort
kubectl apply -f ./snort.yaml

# Flow controller
kubectl apply -f ./flow-controller.yaml

# Malice 
kubectl apply -f ./malice.yaml
```

Check the interfaces inside snort, flow-controller

```sh
kubectl exec -it flow-controller -- ip a
kubectl exec -it snort -- ip a
```

You should be able to ping both interfaces

```sh
kubectl exec -it flow-controller -- ip a
# Ping eth0 of alice
kubectl exec -it snort -- ping XXXXXXXX
# Ping net1 of alice
kubectl exec -it snort -- ping XXXXXXXX
```

Find the name of the snort

```sh
kubectl exec -n kube-system -it antrea-agent-XXXX -- ovs-vsctl show | grep snort 

```

Set up port mirroring to snort 

```sh
kubectl exec -n kube-system -it antrea-agent-XXXX -- ovs-vsctl \
  -- --id=@p get port snort-XXXX \
  -- --id=@m create mirror name=m0 select-all=true output-port=@p \
  -- set bridge br-int mirrors=@m

```

You should be able to ping net1 interface of flow-controller 

```sh
kubectl exec -it flow-controller -- ip a
# Ping eth0 of alice should fail
kubectl exec -it snort -- ping XXXXXXXX
# Ping net1 of alice still works!
kubectl exec -it snort -- ping XXXXXXXX
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
