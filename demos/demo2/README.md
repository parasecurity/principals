## Link a local docker image (or Dockerfile) to kubernetes

Start the local registy

```sh
docker run -d -p 5000:5000 --restart=always --name registry registry:2

```

Build the image and push it 

```sh
docker build . -t XXXX

docker tag dga:v1.0.0 localhost:5000/dga:v1.0.0

docker push localhost:5000/dga:v1.0.0  

```

Link minikube with the local registry

```
#Add --insecure-registry="192.168.49.1:5000" to minikube start
minikube start \
    --vm-driver=docker \
    --extra-config=kubeadm.pod-network-cidr=172.16.0.0/12 \
    --extra-config=kubelet.network-plugin=cni \
    --insecure-registry="192.168.49.1:5000"

```

Add to kubernetes yaml: 192.168.49.1:5000/XXXX

## Create a network traffic pod with 2 network interfaces

Create the local registry with dga, flow-control docker images hosted there

```sh
# Start the local registy
docker run -d -p 5000:5000 --restart=always --name registry registry:2

```

Create dga image

```sh
# Build the dga docker image
docker build . -t dga:v1.0.0

# Tag it 
docker tag dga:v1.0.0 localhost:5000/dga:v1.0.0

# Upload it to the local registry
docker push localhost:5000/dga:v1.0.0

```

Create flow-control image

```sh
# Build the dga docker image
docker build . -t flow-control:v1.0.0                            

# Tag it 
docker tag flow-control:v1.0.0 localhost:5000/flow-control:v1.0.0

# Upload it to the local registry
docker push localhost:5000/flow-control:v1.0.0 

```

Delete the previous minikube session

```sh
minikube delete

```

Start minikube

```sh
minikube start \
    --vm-driver=docker \
    --extra-config=kubeadm.pod-network-cidr=172.16.0.0/12 \
    --extra-config=kubelet.network-plugin=cni \
    --insecure-registry="192.168.49.1:5000"
```

Download and install Andrea

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

Create new network configuration

```sh 
kubectl create -f ./2_port.yaml

```

Create dga, flow-control pods

```sh
# Create dga pod 
kubectl apply -f ./dga.yaml

# Create flow controller pod
kubectl apply -f ./flow-control.yaml

```

Create some alice, malice pods to generate trafic

```sh
# Create alice pod 
kubectl apply -f ./nslookup.yaml

# Create malice pod 
kubectl apply -f ./malice.yaml

```

Check the interfaces inside dga, flow-controller and nslookup

```sh
kubectl exec -it dga -- ip a
kubectl exec -it flow-control -- ip a
kubectl exec -it nslookup -- ip a
```

You should be able to ping both interfaces

```sh
# Get both ips of flow-control
kubectl exec -it flow-control -- ip a

# Ping eth0 of flow-controller 
kubectl exec -it dga -- ping XXXXXXXX

# Ping net1 of flow-controller 
kubectl exec -it dga -- ping XXXXXXXX

```

Find the name of the dga

```sh
kubectl exec -n kube-system -it antrea-agent-XXXX -- ovs-vsctl show | grep dga 

```

Set up port mirroring to snort 

```sh
kubectl exec -n kube-system -it antrea-agent-XXXX -- ovs-vsctl \
  -- --id=@p get port dga-XXXX \
  -- --id=@m create mirror name=m0 select-all=true output-port=@p \
  -- set bridge br-int mirrors=@m

```

Copy antrea_agent_server.py script inside antrea-agent
```sh
# Update antrea-agent
kubectl exec -n kube-system -it antrea-agent-XXXX -- apt-get update

# Install vim
kubectl exec -n kube-system -it antrea-agent-XXXX -- apt-get install vim

# Create server.py file
kubectl exec -n kube-system -it antrea-agent-XXXX -- vim home/server.py
```

Start antrea_agent_server script
```sh
kubectl exec -n kube-system -it antrea-agent-XXXX -- python3 home/server.py

```

Start dga script 
```sh
kubectl exec -it dga -- bash -c "cd tmp && python3 monitory.py -m dga.model"

```

Start flow-control script
```sh
kubectl exec -it flow-control -- python3 forward.py 

```

Send some requests from alice 
```sh
kubectl exec -it alice -- nslookup google.com

kubectl exec -it alice -- nslookup amazon.com

kubectl exec -it alice -- nslookup facebook.com

# All those requests should be printed to dga monitor
```

Send some bad request from malice to be blocked
```sh
kubectl exec -it malice -- nslookup alflaaeoeroagoakrgaorbamfemb.com

```