#!/bin/bash

#---------------------------------------------------------------
#---------------------------------------------------------------
GREEN='\x1b[32m'
BLUE='\x1b[34m'
RED='\x1B[31m'
NC='\033[0m'

bold=$(tput bold)
NORMAL=$(tput sgr0)


/bin/bash ./deploy.sh zoomv3 zoomv3 1 1 1 1 gnbsim
echo "Press Enter to continue..."
read


/bin/bash ./undeploy.sh 1
echo "Press Enter to continue..."
read

/bin/bash ./deploy.sh zoomv3 zoomv3 1 1 1 1
echo "Press Enter to continue..."
read

/bin/bash ./deploy.sh zoomv3 zoomv3 1 1 1 1 gnbsim_only
echo "Press Enter to continue..."
read

/bin/bash ./undeploy.sh 1
echo "Press Enter to continue..."
read


echo "-------------------------------------------------"
echo "Experiment Finished for All Use Cases"
echo $rt
echo "-------------------------------------------------"
