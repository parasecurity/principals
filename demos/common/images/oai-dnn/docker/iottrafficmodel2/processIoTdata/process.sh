#!/bin/bash

rm processedin.txt
rm processedout.txt

touch processedin.txt
touch processedout.txt

function abs_diff {
        echo $(($1 >= $2 ? $1 - $2 : $2 - $1))
}

function ms_diff {
        prev=$1
        curr=$2

        milval=$((1*(10#${curr:9:1}-10#${prev:9:1})))
        secval=$((10*(10#${curr:6:2}-10#${prev:6:2})))
        minval=$((600*(10#${curr:3:2}-10#${prev:3:2})))
        hrval=$((3600*(10#${curr:0:2}-10#${prev:0:2})))

        totaldiff=$(($hrval+$minval+$secval+$milval))

        echo $totaldiff
}

IFS=$'\n'

file=$(cat $1)

SRC="hubv2"

curIUin=0
curIUout=0

dataUsageIn=0
datausageOut=0

prevT=$(echo $file | awk 'FNR == 1 {print $1}')

re='^[0-9]+$'

for i in $file
do

        packettype=$(echo $i | awk '{print $2}' )

        #if [[ $packettype == "IP" ]];
        #then
                currentSRC=$(echo $i | awk '{print $3}' )
                curT=$(echo $i | awk '{print $1}' )
                curU=$(echo $i | awk '{print $NF}')

                if [[ $currentSRC  == *$SRC* ]];
                then
                        if ! [[ $curU =~ $re ]] ;
                        then
                                #echo "OpenWRT packet. Skipping parse"
                                echo ""

                        else
                                curIUout=$(($curIUout+$curU))
                        fi
                else
                        if ! [[ $curU =~ $re ]] ;
                        then
                                #echo "OpenWRT packet. Skipping parse"
                                echo ""

                        else
                                curIUin=$(($curIUin+$curU))
                        fi
                fi

                if [ ${curT:9:1} -ne ${prevT:9:1} ];
                then

                        timeP=$(ms_diff $prevT $curT)

                        echo $timeP

                        datausageOut=$(($curIUout))
                        datausageIn=$(($curIUin))
                        echo $datausageOut >> processedout.txt
                        echo $datausageIn >> processedin.txt
                        curIUout=0
                        curIUin=0

                        for ((j=1;j<=timeP;j++))
                        do
                                echo $curIUout >> processedout.txt
                                echo $curIUin >> processedin.txt
                        done
                fi
                prevT=$curT

        #fi
done

