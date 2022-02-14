#!/bin/bash

	numberofusers=20
        normusers=8
        maliusers=$(($numberofusers-$normusers))

	for ((i=1;i<=$normusers;i++))
        do
            ci=$(($i+9))
            declare GNBSIM_NORM_$i=$(kubectl get pods -n oai  | grep gnbsim$ci | awk '{print $1}')
        done
        for ((i=1;i<=$maliusers;i++))
        do
            ci=$(($i+9+$normusers))
            declare GNBSIM_ATTC_$i=$(kubectl get pods -n oai  | grep gnbsim$ci | awk '{print $1}')
        done

	 
