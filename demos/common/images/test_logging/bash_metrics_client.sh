#!/bin/sh

METRICS_SERVER_POD=$(kubectl get -A po -o wide | grep metrics-client-py | awk '{print $2}'i)
kubectl exec -it -n security $METRICS_SERVER_POD -- bash 
