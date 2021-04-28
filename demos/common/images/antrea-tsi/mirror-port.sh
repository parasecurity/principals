#!/usr/bin/env bash
#
#  Port mirroring script
#
readonly port=$(ovs-vsctl show | grep -o "$NAME[^ ]*" | head -1 )
ovs-vsctl \
  -- --id=@p get port $port\
  -- --id=@m create mirror name=$port-m0 select-all=true output-port=@p \
  -- set bridge br-int mirrors=@m
