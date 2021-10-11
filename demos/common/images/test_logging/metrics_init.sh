#!/usr/bin/env bash
#
# setup of metrics logging server
# Status: on progress, need to be integrated to 
# init.sh after development
#
set -euo pipefail

setupMetricsImages()
{

	# metrics-server
	docker build metrics-server -t metrics-server:common 
	docker tag metrics-server:common localhost:5000/metrics-server:common 
	docker push localhost:5000/metrics-server:common 
	docker rmi localhost:5000/metrics-server:common

	# metrics-client-go
	docker build metrics-client-go -t metrics-client-go:common 
	docker tag metrics-client-go:common localhost:5000/metrics-client-go:common 
	docker push localhost:5000/metrics-client-go:common 
	docker rmi localhost:5000/metrics-client-go:common

	# metrics-client-py
	docker build metrics-client-py -t metrics-client-py:common 
	docker tag metrics-client-py:common localhost:5000/metrics-client-py:common 
	docker push localhost:5000/metrics-client-py:common 
	docker rmi localhost:5000/metrics-client-py:common

}

echo creating metrics server 
setupMetricsImages

echo finished
