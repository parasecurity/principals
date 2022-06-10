#!/bin/bash

IPS=`cat $HOME/tsi/parser.log | grep -E 'applied' | awk '{print $6}' | sort -u | xargs echo | sed 's/ /|/g'`
kubectl get -A po -owide 2> /dev/null | grep -E "$IPS" | awk '{print $7 "\t" $2}'
