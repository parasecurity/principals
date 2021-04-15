#!/usr/bin/env bash
#
#  This services connects to remote syslog server 
#
readonly IP=$1
readonly PORT=$2
 
tail -f var/log/snort/alert 2>&1 | logger -s -n $IP -P $PORT 2>&1