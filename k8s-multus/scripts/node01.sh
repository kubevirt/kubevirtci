#!/bin/bash

set -ex

# Wait for the network to really came up

while [[ `systemctl status docker | grep active | wc -l` -eq 0 ]]
do
 sleep 2
done

kubeadm init --config /etc/kubernetes/kubeadm.conf

kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /etc/kubernetes/multus.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /etc/kubernetes/flannel.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /etc/kubernetes/macvlan-conf.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /etc/kubernetes/ptp-conf.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /etc/kubernetes/bridge-conf.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/master:NoSchedule- 