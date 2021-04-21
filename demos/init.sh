#!/usr/bin/env bash
#
#  Init Script for all demos
#
set -euo pipefail

source common/scripts/funcs.sh

setupCommonImages()
{
	#antrea-tsi
	docker build common/images/antrea-tsi -t antrea-tsi:v1.0.0 &> /dev/null
	docker tag antrea-tsi:v1.0.0 localhost:5000/antrea-tsi:v1.0.0 &> /dev/null
	docker push localhost:5000/antrea-tsi:v1.0.0 &> /dev/null
	docker rmi localhost:5000/antrea-tsi:v1.0.0 &> /dev/null

	#dga
	docker build common/images/dga -t tsi-dga:common &> /dev/null
	docker tag tsi-dga:common localhost:5000/tsi-dga:common &> /dev/null
	docker push localhost:5000/tsi-dga:common &> /dev/null
	docker rmi localhost:5000/tsi-dga:common &> /dev/null

	#flow-control
	docker build common/images/flow-control -t tsi-flow-control:common &> /dev/null
	docker tag tsi-flow-control:common localhost:5000/tsi-flow-control:common &> /dev/null
	docker push localhost:5000/tsi-flow-control:common &> /dev/null
	docker rmi localhost:5000/tsi-flow-control:common &> /dev/null

	if [[ "$CLEAN" == "True" ]]
	then
		docker rmi antrea-tsi:v1.0.0 &> /dev/null
		docker rmi tsi-dga:common &> /dev/null
		docker rmi tsi-flow-control:common &> /dev/null
	fi
}

readonly REGISTRY_STATUS=$(docker ps -q -f "name=minikube_registry")

if [[ "$REGISTRY_STATUS" == "" ]];
then
	[[ "$(docker ps -aq -f "name=minikube_registry")" != "" ]] && docker rm minikube_registry

	msg "Starting local docker registry"
	docker run -d -p 5000:5000 --restart=always --name minikube_registry registry:2
fi

if [ $# -ne 0 ] &&  [ "$1" == "clean" ]
then
	CLEAN=True
else
	CLEAN=False
fi

msg "Creating common images"
setupCommonImages

msg "Finished creating images"
