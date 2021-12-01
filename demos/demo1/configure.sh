#!/usr/bin/env bash

source conf/demo.conf
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
		awk "$awk_script" conf/$file.yaml.conf > yamls/$file.yaml
	done;
}

clean() {
	for file in $KUBE_DEPLOYMENTS; do
		rm yamls/$file.yaml 2> /dev/null
	done;
}

usage() {
	echo "Usage: $1"
	echo "       $1 [OPTION]"
	echo "  Generates yaml deployment files for ddos demo"
	echo ""
	echo ""
	echo "Options:"
	echo "  --clean             remove generated yaml files"
	echo "  --file FILENAME     generate for specific file"
	echo "  --help              print this message"

	exit $2
}

if [ $# -eq 1 ]; then
	if [ $1 = "--clean" ]; then
		clean
	elif [ $1 = "--help" ]; then
		usage $0 0
	else
		usage $0 1
	fi
elif [ $# -eq 0 ]; then
	clean
	generate
elif [ $1 = "--file" ]; then
	shift
	KUBE_DEPLOYMENTS="$@"
	clean
	generate
else
	usage $0 1
fi
