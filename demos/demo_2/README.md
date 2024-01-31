# OAI 5GC and RAN Deployment in K8s

This branch includes the latest deployment integrated with 5G-STREAM. Enabling "proxy" from the helm charts of the VNFs will enable them to peak with the 5G-STREAM Service Communication Proxy. When disabled, 5G deployment will proceed as usual. 

## Deploy
From within the /charts directory, navigate to the helm templates of each of the following components -- "oai-nrf" , "oai-amf" , "oai-smf" , "oai-spgwu-tiny" -- and change the multus interfaces to the one on your local machine. 

Create the oai namespace.
```
kubectl create ns oai
```

From the oai-5gcn/charts folder execute the following,
```
$ helm install <name-of-mysql-deployment> mysql/ -n oai
```
which will first deploy the MYSQL pod.
```
$ ./run.sh
```
The run.sh script will deploy the 5G core. Using the inputs to the script (detailed within the script), the number of slices and users can be adjusted.

## Information about the 5G components

- gnbsim10: gNB stands for gNodeB, the 5G equivalent of 4G's eNodeB. This name suggests a simulator for testing 5G wireless networks.
- oai-amf10: AMF stands for Access and Mobility Management Function. This component is part of the core network that handles aspects related to mobility, session management, and authentication.
- oai-ausf10: AUSF stands for Authentication Server Function. It's responsible for handling the authentication process between the user equipment (UE) and the network.
- oai-dnn10: DNN stands for Data Network Name. It represents the network that UEs connect to for actual network services.
- oai-nrf10: NRF stands for Network Repository Function. It handles the management and storage of information about NF (Network Functions) and services.
- oai-smf10: SMF stands for Session Management Function. It manages users' session contexts and establishes, modifies, and releases sessions.
- oai-spgwu-tiny10: SPGWU stands for Serving Gateway and PDN Gateway in User plane. It's the data forwarding node of the network.
- oai-udm10: UDM stands for Unified Data Management. It manages user data such as subscription data, credentials, and policies.
- oai-udr10: UDR stands for Unified Data Repository. It is a network function that stores data from other network functions.
