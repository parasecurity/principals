#!/bin/sh

ls ../metrics-client-go/cli_* | sed "s/\(.*\)go\/\(.*\).go/\1go\/\2.go results_wl$1_\2_server_v2.txt $1/" | xargs -n3 sh -c './scratch.py -c $1 -o $2 -w $3 -l \"../metrics-server/metrics.log\"' sh
