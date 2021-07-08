#!/usr/bin/env bash
#
#   Generate root/server certificate
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

prerequisites() {

    msg "Checking if openssl is installed"
    PKG_OK=$(dpkg-query -W --showformat='${Status}\n' openssl | grep "install ok installed")
    if [ "" = "$PKG_OK" ]; then
        errmsg "No openssl. Please install openssl."
        exit
    fi

    msg "Checking if internal directory exists"
    if [ ! -d "../../internal/" ] 
    then
        mkdir ../../internal/
    fi
}

generate() {    
    msg " >>>>>>>>>>>>>>>>>> Generate root certificate <<<<<<<<<<<<<<<<<<<<<<"
    msg "Generating root certificate private key: ca.key"
    openssl genrsa -out ca.key 2048

    msg "Generating self-signed root certificate: ca.crt"
    openssl req -new -key ca.key -x509 -days 3650 -out ca.crt -subj /C=CN/ST=Greece/O="tsi"/CN="Root"

    msg " >>>>>>>>>>>>>>>>>> Generate server certificate <<<<<<<<<<<<<<<<<<<<<<"
    msg "Generating server certificate private key: server.key"
    openssl genrsa -out server.key 2048

    msg "Generating server certificate request: server.csr"
    openssl req -new -nodes -key server.key -out server.csr -subj /C=CN/ST=Greece/L=Athens/O="Server"/CN=localhost

    msg "Generating server certificate: server.crt"
    openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt

    msg " >>>>>>>>>>>>>>>>>> Generate client certificate <<<<<<<<<<<<<<<<<<<<<<"
    msg "Generating client certificate private key: client.key"
    openssl genrsa -out client.key 2048

    msg "Generating client certificate request: client.csr"
    openssl req -new -nodes -key client.key -out client.csr -subj /C=CN/ST=Greece/L=Athens/O="Client"/CN=localhost

    msg "Signing client certificate: client.crt"
    openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt

    mv *.key *.csr *.crt ca.srl ../../internal/ 
}

prerequisites
generate
