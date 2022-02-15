#!/bin/bash

dnnpod=$(kubectl get pods -n oai  | grep oai-dnn | awk '{print $1}')


for ((i=1; i<=999; i++));
do
	echo $i
	kubectl exec -n oai "$dnnpod" -- bash -c "nohup /iottrafficmodel/tcpServerClient/build/tcpclient &"
done
