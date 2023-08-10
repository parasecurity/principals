## DEMO ##


# run 
First run the configure file.

    ./configure.sh

Make sure you have the right IP on the server.

Inside the "extras" folder in the demo1 we have some extra documents for easy of deployment

# SCP
From the client run the scp.sh script. Make sure first to have ssh-copy the password for the server, and have the correct server on the script.

    ./scp.sh 

Then run on the same time the following command, on the same director:

    tail -f transfer.log > transfer_time.log

This will log the trasfer speed.

# VOIP
From the client run (-r rate limit):

    sudo su
    ulimit -n 128000
    /usr/local/bin/sipp 147.27.39.38:5063 -sf /tmp/uac_pcap.xml -r 3000 -rp 1000 -trace_stat -fd 1 -i 147.27.39.37 -l -1

From the server run:

    /usr/local/bin/sipp -sn uas -i 147.27.39.38 -p 5063

To monitor bandwith :

    sudo iftop -ni eno8303

# Time
The time.sh script converts the time to readable.

    ./time.sh <time_in_milliseconds>

# Update flow servers
During every run, execute the update_flow_servers script to be sure that no errors will occur.

    ./update_flow_servers.sh

# Bandwith 
To calculate the packes sent from alice do:

    cat tsi/tsi.log | grep alice > logs_default.log
    ./bandwith.sh > bandwith.log
     
If you want to extract the success per second only :

    cat bandwith.log | grep Succ | awk '{print $2}' | cut -d ',' -f 1


