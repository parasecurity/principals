#!/bin/sh

for res in $@; do

	wl=`echo $res | cut -f2 -d_ | cut -f2 -dl`
	t=$(( 10*$wl + 2*$wl ))

	echo $(tput bold)$(tput setaf 2)==========================================================
	echo $res$(tput sgr0)
	grep total $res | grep -v $t

	tail -6 $res
done
