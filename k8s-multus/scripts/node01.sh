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

kubeadm init --config /etc/kubernetes/kubeadm.conf

cd multus-cni/examples

kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f clusterrole.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f crd.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f flannel-conf.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f macvlan-conf.yml

kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/intel/multus-cni/dev/network-plumbing-working-group-crd-change/examples/multus-with-flannel.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/master:NoSchedule- 