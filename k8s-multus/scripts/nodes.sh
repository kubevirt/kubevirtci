#!/bin/bash

set -ex

# Wait for the network to really came up
while [[ `cat /proc/sys/net/ipv4/ip_forward` -eq 0 ]]
do
 sleep 2
done

while [[ ! -f /proc/sys/net/bridge/bridge-nf-call-iptables ]]
do
 sleep 2
done

kubeadm join --token abcdef.1234567890123456 192.168.66.101:6443 --ignore-preflight-errors=all --discovery-token-unsafe-skip-ca-verification=true
