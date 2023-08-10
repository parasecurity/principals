#!/bin/bash

TIMESTAMP_MICRO=$1

SECONDS=$(echo "$TIMESTAMP_MICRO / 1000000" | bc)
MICROSECONDS=$(echo "$TIMESTAMP_MICRO % 1000000" | bc)

DATE=$(date -u -d @$SECONDS +"%Y-%m-%d %H:%M:%S")

echo "${DATE}.${MICROSECONDS}"
