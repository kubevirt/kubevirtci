#!/bin/bash

set -ex

# Wait for docker, else network might not be ready yet
while [[ `systemctl status docker | grep active | wc -l` -eq 0 ]]
do
    sleep 2
done

kubeadm init --config /etc/kubernetes/kubeadm.conf
version=`kubectl version --short --client | cut -d":" -f2 |sed  's/ //g' | cut -c2- | cut -d"." -f2`

if [[ ${version} -ge "12" ]]; then
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/flannel-ge-12.yaml
else
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/flannel.yaml
fi

kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/kubernetes-multus.yaml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/multus.yaml

kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/cni-plugins-ds.yaml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/kubernetes-ovs-cni.yaml

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
