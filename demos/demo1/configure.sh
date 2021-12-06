#!/usr/bin/env bash

source ../common/scripts/funcs.sh

CONF_DIR=conf
ROOT_DIR=`git rev-parse --show-toplevel`
DEMO_DIR="$ROOT_DIR/demos/demo1"
source $CONF_DIR/demo.conf
REGISTRY=$REGISTRY_IP:$REGISTRY_PORT

awk_script=" {
	sub(\"%ANTREA_TSI%\", \"$REGISTRY/$ANTREA_TSI\", \$0);
	sub(\"%TSI_LOGGING%\", \"$REGISTRY/$TSI_LOGGING\", \$0);
	sub(\"%TSI_API%\", \"$REGISTRY/$TSI_API\", \$0);
	sub(\"%REGISTRY%\", \"$REGISTRY\", \$0);
	sub(\"%REGISTRY_IP%\", \"$REGISTRY_IP\", \$0);
	print
} "


generate() {
	mkdir -p yamls/security
	mkdir -p yamls/pods
	for file in $KUBE_DEPLOYMENTS; do
		msg "Generating yamls/$file.yaml"
		awk "$awk_script" $CONF_DIR/$file.yaml.conf > yamls/$file.yaml
	done;
}

check() {
	if [ ! -d $CONF_DIR ]; then
		errmsg "Error: Directory $CONF_DIR does not exist!"
		wrnmsg "  Please make sure you are in $DEMO_DIR directory and\n"
		wrnmsg "  your repo is up to date.\n"
		exit 2
	fi
	for file in $KUBE_DEPLOYMENTS; do
		if [ ! -f $CONF_DIR/$file.yaml.conf ]; then
			errmsg "Error: file $CONF_DIR/$file.yaml.conf not found!"
			exit 2
		fi
	done;
}

# this funcrion cleans all yaml files generated from this script
# that where secified in demo.conf. If you want to clean files 
# generated with --file option you should remove them manualy,
# or add them in KUBE_DEPLOYMENTS in demo.conf
clean() {
	for file in $KUBE_DEPLOYMENTS; do
		if [ ! -f yamls/$file.yaml ]; then
			continue
		fi
		rm yamls/$file.yaml 2> /dev/null
	done;

	if [ -d yamls/security ]; then
		if [ ! "$(ls yamls/security 2> /dev/null)" ] ; then
			msg "cleaning security"
			rm -rf yamls/security 2> /dev/null
		fi
	fi
	if [ -d yamls/pods ]; then
		if [ ! "$(ls yamls/pods 2> /dev/null)" ] ; then
			msg "cleaning pods"
			rm -rf yamls/pods 2> /dev/null
		fi
	fi
	if [ -d yamls ]; then
		if [ ! "$(ls yamls 2> /dev/null)" ] ; then
			msg "removing yamls"
			rm -rf yamls 2> /dev/null
		fi
	fi
}

usage() {
	echo "Usage: $1"
	echo "       $1 [OPTION]"
	echo "  Generates yaml deployment files for ddos demo"
	echo ""
	echo "Options:"
	echo "  --clean             remove generated yaml files. Files not defined in "
	echo "                      KUBE_DEPLOYMENTS will NOT be removed"
	echo "  --file FILENAME     generate for specified file FILENAME sould be in a form"
	echo "                      <pods|security|other_custom_floder>/<filename_without_extention>"
	echo "                      e.g.: ./configure --file pods/alice"
	echo "                      If a custom file is used, it should have .yaml.conf extention"
	echo "  --help              print this message"

	exit $2
}

check
if [ $# -eq 1 ]; then
	if [ $1 = "--clean" ]; then
		clean
	elif [ $1 = "--help" ]; then
		usage $0 0
	else
		usage $0 1
	fi
elif [ $# -eq 0 ]; then
	generate
elif [ $1 = "--file" ]; then
	shift
	KUBE_DEPLOYMENTS="$@"
	generate
else
	usage $0 1
fi
