#!/bin/bash

# Start the first process
python3 -m http.server &
  
# Start the second process
/app/dns_stitcher live eth0
  
