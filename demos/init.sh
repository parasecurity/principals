#!/usr/bin/env bash
#
#  Init Script for all demos
#
set -euo pipefail

source common/scripts/funcs.sh

setupCommonImages()
{
	#antrea-tsi
	cd common/images/antrea-tsi
	make
	make push
	cd ../../../

	#tsi-tools
	#cd common/images/tsi-tools
	#make
	#make push
	#cd ../../../

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

        #analyser
	docker build common/images/analyser -t tsi-analyser:common &> /dev/null
	docker tag tsi-analyser:common localhost:5000/tsi-analyser:common &> /dev/null
	docker push localhost:5000/tsi-analyser:common &> /dev/null
	docker rmi localhost:5000/tsi-analyser:common &> /dev/null

	if [[ "$CLEAN" == "True" ]]
	then
		docker rmi antrea-tsi:v1.0.0 &> /dev/null
		docker rmi tsi-dga:common &> /dev/null
		docker rmi tsi-flow-control:common &> /dev/null
		docker rmi tsi-analyser:common &> /dev/null
	fi
}

setupDemo7Images()
{
	docker build demo7/images/flow_control_server -t tsi-flow-server:demo7 &> /dev/null
	docker tag tsi-flow-server:demo7 localhost:5000/tsi-flow-server:demo7 &> /dev/null
	docker push localhost:5000/tsi-flow-server:demo7 &> /dev/null
	docker rmi localhost:5000/tsi-flow-server:demo7 &> /dev/null

	if [[ "$CLEAN" == "True" ]]
	then
		docker rmi tsi-flow-server:demo7 &> /dev/null
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

msg "Creating demo 7 images"
setupDemo7Images

msg "Finished creating images"
