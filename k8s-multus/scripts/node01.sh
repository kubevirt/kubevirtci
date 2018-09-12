#!/bin/bash

set -ex

# Wait for docker, else network might not be ready yet
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

# Wait for api server to be up.
kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --no-headers
kubectl_rc=$?
retry_counter=0
while [[ $retry_counter -lt 20 && $kubectl_rc -ne 0 ]]; do
    sleep 10
    echo "Waiting for api server to be available..."
    kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --no-headers
    kubectl_rc=$?
    retry_counter=$((retry_counter + 1))
done
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/local-volume.yaml
