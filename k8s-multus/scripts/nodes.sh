#!/bin/bash

set -ex

# Wait for the network to really came up
while [[ `systemctl status docker | grep active | wc -l` -eq 0 ]]
do
 sleep 2
done

kubeadm join --token abcdef.1234567890123456 192.168.66.101:6443 --ignore-preflight-errors=all --discovery-token-unsafe-skip-ca-verification=true
