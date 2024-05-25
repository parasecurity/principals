#!/bin/bash

#Get the list of pods
pods=$(kubectl get pods -n security  -o jsonpath='{.items[*].metadata.name}'|  tr " " "\n" | grep 'flow-server')

#Loop over the pods and run the command
for pod in $pods; do
  echo "Running 'apt-get update' on $pod"
  # Keep trying until the command succeeds
  until kubectl exec -it $pod -n security -- /bin/bash -c "apt-get update"; do
    echo "Command failed, retrying in 2 seconds..."
    sleep 2
  done
done
