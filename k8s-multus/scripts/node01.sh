#!/bin/bash

set -ex

# Wait for the network to really came up

while [[ `systemctl status docker | grep active | wc -l` -eq 0 ]]
do
 sleep 2
done

kubeadm init --config /etc/kubernetes/kubeadm.conf

cd multus-cni/examples

kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f clusterrole.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f crd.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f flannel-conf.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f macvlan-conf.yml

kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /etc/kubernetes/cni.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/master:NoSchedule- 