#!/bin/bash

startTime=0
successCount=0
failCount=0
timeCounter=0

while IFS= read -r line; do
    current_time=$(echo $line | cut -d ' ' -f4)
    status=$(echo $line | cut -d ' ' -f5)

    SECONDS=$(echo "$current_time / 1000000" | bc )
    DATE=$(date -u -d @$SECONDS +"%Y-%m-%d %H:%M:%S")
    current_seconds=$SECONDS

    # Check if start time is not set yet    
    if [[ $startTime -eq 0 ]]; then
        startTime=$current_seconds
    fi

    # Check if more than a second has passed
    if [[ $current_seconds -gt $startTime ]]; then
        echo $timeCounter $DATE
        echo "Successes: $successCount, Failures: $failCount"
        successCount=0
        failCount=0
        startTime=$current_seconds
        ((timeCounter++))

    fi

    # Check if the request was successful or not
    if [[ $status == "Response" ]]; then
        ((successCount++))
    else
        ((failCount++))
    fi
done < "logs_default.log"

# Print counts for the last period
echo $timeCounter $DATE
echo "Successes: $successCount, Failures: $failCount"

