#!/usr/bin/env bash
#
#  Port mirroring script
#
readonly DGA=$(ovs-vsctl show | grep -o "dga[^ ]*" | head -1 )
ovs-vsctl \
  -- --id=@p get port $DGA \
  -- --id=@m create mirror name=$DGA-m0 select-all=true output-port=@p \
  -- set bridge br-int mirrors=@m