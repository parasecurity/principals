#!/usr/bin/env bash
#
#  Script to create a local docker registry
#
set -euo pipefail

readonly REGISTRY_STATUS=$(docker ps -q -f "name=local_registry")

if [[ "$REGISTRY_STATUS" == "" ]];
then
	[[ "$(docker ps -aq -f "name=local_registry")" != "" ]] && docker rm local_registry

	echo "Starting local docker registry"
	docker run -d -p 5000:5000 --restart=always --name local_registry registry:2
fi
