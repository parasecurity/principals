#!/bin/bash

st=`date +%s`
#---------------------------------------------------------------
#---------------------------------------------------------------
GREEN='\x1b[32m'
BLUE='\x1b[34m'
NC='\033[0m'

bold=$(tput bold)
NORMAL=$(tput sgr0)
#---------------------------------------------------------------
#---------------------------------------------------------------

gnbsimim=$1
dnnim=$2
sa_version=1.1.17 

((user=$3+10))
((trafficusers=$3-1))
((slice=$4+9))
u=10



echo "-------------------------------------------------"
echo -e "${BLUE} ${bold} gNBSIM image set to $gnbsimim ${NC} ${NORMAL}"
echo -e "${BLUE} ${bold} DNN image is set to $dnnim ${NC} ${NORMAL}"
echo "-------------------------------------------------"


echo "-------------------------------------------------"
echo -e "${BLUE} ${bold} Deploying $4 slices with $3 users each ${NC} ${NORMAL}"
echo "-------------------------------------------------"

echo "-------------------------------------------------"
echo -e "${GREEN} ${bold} Starting 5G Core deployment ${NC} ${NORMAL}"
echo "-------------------------------------------------"

for arg in "$@"; do
    if [ "$arg" == "gnbsim" ]; then
        found=true
        break
    fi
done

for arg in "$@"; do
	if [ "$arg" == "gnbsim_only" ]; then
		for ((s=10;s<=$slice;s++))
		do
			amfpod=$(kubectl get pods -n oai  | grep amf$s | awk '{print $1}')
			amfeth0=$(kubectl exec -n oai $amfpod -c amf -- ifconfig | grep "inet 10.8" | awk '{print $2}')
			upfpod=$(kubectl get pods -n oai  | grep spgwu-tiny$s | awk '{print $1}')
			upfeth0=$(kubectl exec -n oai $upfpod -c spgwu -- ifconfig | grep "inet 10.8" | awk '{print $2}')
			((z=$s-9))
				ip=2
				for ((ut=0;ut<$3;ut++))
				do
					#-----------------------------GNBSIM Deployment----------------------------------------
					sed -i "2s/.*/name: gnbsim$u/" gnbsim/Chart.yaml
					sed -i "6s/.*/  version: ${gnbsimim}/" gnbsim/values.yaml
					sed -i "28s/.*/  name: \"gnbsim-sa$u\"/" gnbsim/values.yaml
					sed -i "/ngappeeraddr/c\  ngappeeraddr: \"$amfeth0\"" gnbsim/values.yaml
					sed -i "/gnbid/c\  gnbid: \"$u\"" gnbsim/values.yaml
					sed -i "/msin/c\  msin: \"00000000$u\"" gnbsim/values.yaml
					sed -i "/key/c\  key: \"0C0A34601D4F07677303652C046253$u\"" gnbsim/values.yaml
					sed -i "/sst/c\  sst: \"2$s\"" gnbsim/values.yaml

					helm install gnb$u gnbsim/ -n oai 
					sleep 10
					echo -e "${BLUE} ${bold} GNBSIM$u deployed ${NC} ${NORMAL}"
					gnbsimpod=$(kubectl get pods -n oai  | grep gnbsim$u | awk '{print $1}')
					gnbsimeth0=$(kubectl exec -n oai $gnbsimpod -c gnbsim -- ifconfig | grep "inet 10.8" | awk '{print $2}')


					#-----------------------------DNN Deployment-------------------------------------------
					sed -i "4s/.*/  name: oai-dnn$u/" oai-dnn/02_deployment.yaml
					sed -i "6s/.*/    app: oai-dnn$u/" oai-dnn/02_deployment.yaml
					sed -i "11s/.*/      app: oai-dnn$u/" oai-dnn/02_deployment.yaml
					sed -i "17s/.*/        app: oai-dnn$u/" oai-dnn/02_deployment.yaml

					kubectl apply -k oai-dnn/
					sleep 2
					echo -e "${BLUE} ${bold} DNN$u deployed ${NC} ${NORMAL}"
					dnnpod=$(kubectl get pods -n oai  | grep oai-dnn$u | awk '{print $1}')
					dnneth0=$(kubectl exec -n oai $dnnpod -- ifconfig | grep "inet 10.8" | awk '{print $2}')
					
					kubectl exec -it -n oai $dnnpod -- iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
					kubectl exec -it -n oai $dnnpod -- ip route add 12.1.1.0/24 via $upfeth0 dev eth0
					kubectl exec -it -n oai $gnbsimpod -c gnbsim -- ip route replace $dnneth0 via 0.0.0.0 dev eth0 src 12.1.1.$ip
					((ip+=1))
					((u+=1))
					echo "-------------------------------------------------"
				done
		done
		exit 0
	fi
done

for ((s=10;s<=$slice;s++))
do
	((z=$s-9))
	#------------------------NRF-------------------------
	sed -i "18s/.*/  version: $sa_version/" oai-nrf/values.yaml
	sed -i "22s/.*/name: oai-nrf$s/" oai-nrf/Chart.yaml
	sed -i "30s/.*/  name: \"oai-nrf$s-sa\"/" oai-nrf/values.yaml
	helm install nrf$s oai-nrf/ -n oai
	sleep 6
	echo -e "${GREEN} ${bold} NRF$s deployed ${NC} ${NORMAL}"

	#------------------------UDR-------------------------
	sed -i "17s/.*/  version: $sa_version/" oai-udr/values.yaml
	sed -i "23s/.*/name: oai-udr$s/" oai-udr/Chart.yaml
	sed -i "/nrfFqdn/c\  nrfFqdn: \"oai-nrf$s-svc\"" oai-udr/values.yaml
	sed -i "30s/.*/  name: \"oai-udr$s-sa\"/" oai-udr/values.yaml
	helm install udr$s oai-udr/ -n oai
	sleep 6
	echo -e "${GREEN} ${bold} UDR$s deployed ${NC} ${NORMAL}"
	
	#------------------------UDM-------------------------
	sed -i "17s/.*/  version: $sa_version/" oai-udm/values.yaml
	sed -i "23s/.*/name: oai-udm$s/" oai-udm/Chart.yaml
	sed -i "/nrfFqdn/c\  nrfFqdn: \"oai-nrf$s-svc\"" oai-udm/values.yaml
	sed -i "/udrFqdn/c\  udrFqdn: \"oai-udr$s-svc\"" oai-udm/values.yaml
	sed -i "29s/.*/  name: \"oai-udm$s-sa\"/" oai-udm/values.yaml
	helm install udm$s oai-udm/ -n oai
	sleep 6
	echo -e "${GREEN} ${bold} UDM$s deployed ${NC} ${NORMAL}"
	
	#------------------------AUSF------------------------
	sed -i "17s/.*/  version: $sa_version/" oai-ausf/values.yaml
	sed -i "22s/.*/name: oai-ausf$s/" oai-ausf/Chart.yaml
	sed -i "/nrfFqdn/c\  nrfFqdn: \"oai-nrf$s-svc\"" oai-ausf/values.yaml
	sed -i "/udmFqdn/c\  udmFqdn: \"oai-udm$s-svc\"" oai-ausf/values.yaml
	sed -i "31s/.*/  name: \"oai-ausf$s-sa\"/" oai-ausf/values.yaml
	helm install ausf$s oai-ausf/ -n oai
	sleep 6
	echo -e "${GREEN} ${bold} AUSF$s deployed ${NC} ${NORMAL}"
	
	#------------------------AMF-------------------------
	sed -i "22s/.*/name: oai-amf$s/" oai-amf/Chart.yaml	
	sed -i "17s/.*/  version: $sa_version/" oai-amf/values.yaml
	sed -i "/nrfFqdn/c\  nrfFqdn: \"oai-nrf$s-svc\"" oai-amf/values.yaml
	sed -i "/smfFqdn/c\  nrfFqdn: \"oai-smf$s-svc\"" oai-amf/values.yaml
	sed -i "/ausfFqdn/c\  ausfFqdn: \"oai-ausf$s-svc\"" oai-amf/values.yaml
	sed -i "29s/.*/  name: \"oai-amf$s-sa\"/" oai-amf/values.yaml
	sed -i "/sst0/c\  sst0: \"2$s\"" oai-amf/values.yaml
	helm install amf$s oai-amf/ -n oai
	sleep 6
	echo -e "${GREEN} ${bold} AMF$s deployed ${NC} ${NORMAL}"
	amfpod=$(kubectl get pods -n oai  | grep amf$s | awk '{print $1}')
	amfeth0=$(kubectl exec -n oai $amfpod -c amf -- ifconfig | grep "inet 10.8" | awk '{print $2}')
	
	#------------------------SMF-------------------------
	sed -i "22s/.*/name: oai-smf$s/" oai-smf/Chart.yaml
	sed -i "17s/.*/  version: $sa_version/" oai-smf/values.yaml
	sed -i "/nrfFqdn/c\  nrfFqdn: \"oai-nrf$s-svc\"" oai-smf/values.yaml
	sed -i "/udmFqdn/c\  udmFqdn: \"oai-udm$s-svc\"" oai-smf/values.yaml
	sed -i "/amfFqdn/c\  amfFqdn: \"oai-amf$s-svc\"" oai-smf/values.yaml
	sed -i "29s/.*/  name: \"oai-smf$s-sa\"/" oai-smf/values.yaml
	sed -i "/nssaiSst0/c\  nssaiSst0: \"2$s\"" oai-smf/values.yaml
	helm install smf$s oai-smf/ -n oai
	sleep 6
	echo -e "${GREEN} ${bold} SMF$s deployed ${NC} ${NORMAL}"
	
	#------------------------UPF-------------------------
	sed -i "22s/.*/name: oai-spgwu-tiny$s/" oai-spgwu-tiny/Chart.yaml
	sed -i "/nrfFqdn/c\  nrfFqdn: \"oai-nrf$s-svc\"" oai-spgwu-tiny/values.yaml
	sed -i "/fqdn/c\  fqdn: \"oai-spgwu-tiny$s-svc\"" oai-spgwu-tiny/values.yaml
	sed -i "/oai-spgwu-tiny-sa/c\  name: \"oai-spgwu-tiny$s-sa\"" oai-spgwu-tiny/values.yaml
	sed -i "24s/.*/  name: \"oai-spgwu-tiny$s\"/" oai-spgwu-tiny/values.yaml
	sed -i "/nssaiSst0/c\  nssaiSst0: \"2$s\"" oai-spgwu-tiny/values.yaml
	helm install upf$s oai-spgwu-tiny/ -n oai
	sleep 6
	echo -e "${GREEN} ${bold} UPF$s deployed ${NC} ${NORMAL}"
	upfpod=$(kubectl get pods -n oai  | grep spgwu-tiny$s | awk '{print $1}')
	upfeth0=$(kubectl exec -n oai $upfpod -c spgwu -- ifconfig | grep "inet 10.8" | awk '{print $2}')
	
	echo "-------------------------------------------------"
	echo -e "${GREEN} ${bold} Finished Core VNF deployment. Starting RAN. ${NC} ${NORMAL}"
	echo "-------------------------------------------------"

	if [ "$found" == true ]; then
    	ip=2
		for ((ut=0;ut<$3;ut++))
		do
			#-----------------------------GNBSIM Deployment----------------------------------------
			sed -i "2s/.*/name: gnbsim$u/" gnbsim/Chart.yaml
			sed -i "6s/.*/  version: ${gnbsimim}/" gnbsim/values.yaml
			sed -i "28s/.*/  name: \"gnbsim-sa$u\"/" gnbsim/values.yaml
			sed -i "/ngappeeraddr/c\  ngappeeraddr: \"$amfeth0\"" gnbsim/values.yaml
			sed -i "/gnbid/c\  gnbid: \"$u\"" gnbsim/values.yaml
			sed -i "/msin/c\  msin: \"00000000$u\"" gnbsim/values.yaml
			sed -i "/key/c\  key: \"0C0A34601D4F07677303652C046253$u\"" gnbsim/values.yaml
			sed -i "/sst/c\  sst: \"2$s\"" gnbsim/values.yaml

			helm install gnb$u gnbsim/ -n oai 
			sleep 10
			echo -e "${BLUE} ${bold} GNBSIM$u deployed ${NC} ${NORMAL}"
			gnbsimpod=$(kubectl get pods -n oai  | grep gnbsim$u | awk '{print $1}')
			gnbsimeth0=$(kubectl exec -n oai $gnbsimpod -c gnbsim -- ifconfig | grep "inet 10.8" | awk '{print $2}')


			#-----------------------------DNN Deployment-------------------------------------------
			sed -i "4s/.*/  name: oai-dnn$u/" oai-dnn/02_deployment.yaml
			sed -i "6s/.*/    app: oai-dnn$u/" oai-dnn/02_deployment.yaml
			sed -i "11s/.*/      app: oai-dnn$u/" oai-dnn/02_deployment.yaml
			sed -i "17s/.*/        app: oai-dnn$u/" oai-dnn/02_deployment.yaml

			kubectl apply -k oai-dnn/
			sleep 2
			echo -e "${BLUE} ${bold} DNN$u deployed ${NC} ${NORMAL}"
			dnnpod=$(kubectl get pods -n oai  | grep oai-dnn$u | awk '{print $1}')
			dnneth0=$(kubectl exec -n oai $dnnpod -- ifconfig | grep "inet 10.8" | awk '{print $2}')
			
			kubectl exec -it -n oai $dnnpod -- iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
			kubectl exec -it -n oai $dnnpod -- ip route add 12.1.1.0/24 via $upfeth0 dev eth0
			kubectl exec -it -n oai $gnbsimpod -c gnbsim -- ip route replace $dnneth0 via 0.0.0.0 dev eth0 src 12.1.1.$ip
			((ip+=1))
			((u+=1))
			echo "-------------------------------------------------"
		done
	fi
done


echo "-------------------------------------------------"
echo -e "${GREEN} ${bold} Finished 5G Deployment ${NC} ${NORMAL}"
echo "-------------------------------------------------"