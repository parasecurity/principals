#!/usr/bin/env bash
#
#  Init Script for all demos
#
set -euo pipefail

source common/scripts/funcs.sh

readonly MINIKUBE_STATUS=$(minikube status | grep host | awk '{print $2}')
readonly REGISTRY_STATUS=$(docker ps -q -f "name=minikube_registry")

if [[ "$MINIKUBE_STATUS" == "Running" ]];
then
	if [ $# -ne 0 ] &&  [ "$1" == "force" ]
	then
		minikube stop && minikube delete
	else
		echo "Please first stop/clean minikube before deploying the demo cluster"
		exit 0
	fi
fi

if [[ "$REGISTRY_STATUS" == "" ]];
then
	echo "Please run '../init.sh' first to populate the image registry"
	exit 0
fi

msg "Starting minikube cluster"
minikube start \
    --vm-driver=docker \
    --network-plugin=cni \
	--cni=common/images/antrea-tsi.yml \
    --insecure-registry="192.168.49.1:5000"
	
waitUntilAllPodsRun

msg "Cluster is up"

msg "To stop the cluster run 'minikube stop && minikube delete'"