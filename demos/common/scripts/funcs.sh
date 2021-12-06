#!/usr/bin/env bash
#
#  Useful functions to be sourced
#

msg()
{
	local message="$1"
	local bold=$(tput bold)
	local normal=$(tput sgr0)
	
	local color=$(tput setaf 2)
	local color_default=$(tput setaf 9)

	echo ""
	echo "${bold}${color}${message}${color_default}${normal}"
}

errmsg()
{
	local message="$1"
	local bold=$(tput bold)
	local normal=$(tput sgr0)
	
	local color=$(tput setaf 1)
	local color_default=$(tput setaf 9)

	echo ""
	echo "${bold}${color}${message}${color_default}${normal}"
}

wrnmsg()
{
	local message="$1"
	local bold=$(tput bold)
	local normal=$(tput sgr0)
	
	local color=$(tput setaf 4)
	local color_default=$(tput setaf 9)

	echo -en "${bold}${color}${message}${color_default}${normal}"
}

waitUntilAllPodsRun()
{
	wrnmsg "\tWaiting for all pods to be deployed. This might take a while."

	while [[ "$(kubectl get -A pods --field-selector status.phase!=Running -o name)" != "" ]];
		do
			wrnmsg "."
			sleep 2
		done
	
	while [[ "$(kubectl get -A pods | grep 'Error\|CrashLoopBackOff' )" != "" ]];
        do
                wrnmsg "."
                sleep 2
        done
	echo ""
}
