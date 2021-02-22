#!/bin/bash

tcpdump -l -i eth0 -nn not host 10.0.2.2 and not dst 10.0.2.15 2>&1 | python3 /flow_monitor.py

