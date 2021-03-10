#!/usr/bin/env bash
set -euo pipefail

msg()
{
	local message="$1"
	local bold=$(tput bold)
	local normal=$(tput sgr0)

	echo "${bold}${message}${normal}"
}

waitUntilAllPodsRun()
{
	echo -en "\tWaiting for all pods to be deployed. This might take several minutes."

	while [[ "$(kubectl get -A pods --field-selector status.phase!=Running -o name)" != "" ]];
	do
		echo -n "."
		sleep 10
	done

	echo ""
}

msg "Deleting existing minikube configuration"
minikube stop > /dev/null & true
minikube delete > /dev/null & true

msg "Starting minikube cluster"

minikube start \
    --vm-driver=docker \
    --network-plugin=cni \
    --extra-config=kubeadm.pod-network-cidr=172.16.0.0/16 \
    --extra-config=kubelet.network-plugin=cni \
    --insecure-registry="192.168.49.1:5000"

msg "Adding Antrea CNI"

kubectl apply \
    -f https://github.com/vmware-tanzu/antrea/releases/download/v0.12.0/antrea.yml \
    > /dev/null

waitUntilAllPodsRun

# Setup Antrea Agent alias
readonly ANTREA_AGENT=$(kubectl get -A po | grep "antrea-agent" | awk '{print $2}')
readonly ANTREA_POD="kubectl -n kube-system exec -it $ANTREA_AGENT -c antrea-agent -- "

msg "Adding some demo pods"
kubectl apply -f ./utils/alices.yaml > /dev/null

waitUntilAllPodsRun

echo -e "\n\n"

msg "You should have a working kubernetes cluster running ovs!"
