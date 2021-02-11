#!/bin/bash

sudo usermod -aG docker "$USER"
mkdir .kube
touch .kube/config
chmod 600 .kube/config
