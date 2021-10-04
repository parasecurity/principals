#!/bin/sh

test_list=`ls clis/cli*`

for workload in 1 5 10 50 100; do
	ls clis/cli* |\
	sed "s/clis\/\(.*\).go/clis\/\1.go results_wl${workload}_\1_server_v1.txt ${workload}/" |\
	xargs -n3 sh -c './../scratch.py -c $1 -o $2 -w $3 -l ../../metrics-server/logs.log' sh
done
