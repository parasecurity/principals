#!/usr/bin/env bash
set -euo pipefail

readonly ANTREA_AGENT=$(kubectl get -A po | grep "antrea-agent" | awk '{print $2}')

echo "alias antrea=\"kubectl -n kube-system exec -it $ANTREA_AGENT -c antrea-agent -- \"" | tee -a $HOME/.bash_aliases

