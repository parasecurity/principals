#!/usr/bin/env bash
#
#  Init Script for all demos
#
set -euo pipefail

source common/scripts/funcs.sh

if [ "$EUID" -ne 0 ]; then
        echo "Please run as root"
        exit
fi

setupCommonImages()
{
	#antrea-tsi
	cd common/images/antrea-tsi
	make
	make push
	cd ../../../

	#api
	cd common/images/api
	sudo make
	make push
	cd ../../../

	#dga
	cd common/images/dga
	make
	make push
	cd ../../../

	#analyser
	cd common/images/analyser
	make
	make push
	cd ../../../

	#snort
	cd common/images/snort
	make
	make push
	cd ../../../

}

clean(){
	#antrea-tsi
	cd common/images/antrea-tsi
	make clean && true
	cd ../../../

	#api
	cd common/images/api
	make clean && true
	cd ../../../

	#dga
	cd common/images/dga
	make clean && true
	cd ../../../

	#analyser
	cd common/images/analyser
	make clean && true  
	cd ../../../

	#snort
	cd common/images/snort
	make clean && true
	cd ../../../
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
	clean
	exit
fi

msg "Creating common images"
setupCommonImages

msg "Finished creating images"
