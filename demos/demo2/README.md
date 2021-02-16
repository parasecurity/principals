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
kubectl exec -n kube-system -it antrea-agent-s6gml -- 
  ovs-vsctl -- --id=@p get port snort-0dc353 \
            -- --id=@m create mirror name=m0 select-all=true output-port=@p \
            -- set bridge br-int mirrors=@m
```
## Inspect trafic inside brige

Open snort container and execute snort

```
kubectl exec -it snort -- bash
snort
```

Ping google from malice 

```
kubectl exec -it malice -- ping -c3 google.com

```

Ping and responce should appear on snort!

