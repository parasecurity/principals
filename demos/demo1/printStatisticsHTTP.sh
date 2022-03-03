#!/bin/bash

statistics()
{
  ANTREA_POD=$(kubectl get -A po -o wide | grep "antrea-agent" | head -1 | awk '{print $2}')
  PORT_NUMBERS=$(kubectl exec -n kube-system "$ANTREA_POD" -- bash -c 'ovs-ofctl dump-ports-desc br-int | grep malic ' | awk -F "(" '{print $1}')

  echo "Packages Transmitted"
  for i in {1..12}
    do 
      PORT=$(echo $PORT_NUMBERS | awk  -v i="$i" '{ print $i }')
      NAME=$(kubectl exec -n kube-system "$ANTREA_POD" -- bash -c 'ovs-ofctl dump-ports-desc br-int' | grep $PORT | awk '{ print $1 }' | awk -F "(" '{ print $2 }' | awk -F ")" '{ print $1 }' | awk -F "-" '{ print $1 }') 
      STATS=$(kubectl exec -n kube-system "$ANTREA_POD" -- bash -c 'ovs-ofctl dump-ports br-int' | grep $PORT | awk '{print $4 $5}' | head -n 1)
      IP=$(kubectl get pods -A -o wide | grep "$NAME " | awk '{ print $7 }')
      echo "$NAME $IP: $STATS"
  done

  echo "Packages Dropped"
  kubectl exec -n kube-system "$ANTREA_POD" -- bash -c "ovs-ofctl dump-flows br-int | grep 0x0 | grep nw_src" | awk '{ print $7" "$4" "$5 }'
}

statistics 



