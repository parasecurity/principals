#!/bin/bash

#   tail -f transfer.log > transfer_time.log

# function to be executed when the script is interrupted
cleanup() {
    kill $scp_pid
    exit
}

# run the function cleanup when the script is interrupted
trap cleanup SIGHUP SIGINT SIGTERM

# run the scp command in the background
script -q -c  "scp slyronis@kronos.mhl.tuc.gr:/tmp/hugefile /dev/null" > transfer.log &

# get the process ID of the scp command
scp_pid=$!

while true
do
  echo "$(date)" >> transfer.log
  sleep 1
done

