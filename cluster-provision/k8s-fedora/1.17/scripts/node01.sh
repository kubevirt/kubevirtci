#!/bin/bash

set -ex

source /tmp/scripts/cnis-map.sh

version=`kubectl version --short --client | cut -d":" -f2 |sed  's/ //g' | cut -c2- `
minor_version=`echo $version | cut -d"." -f2 `
cni_manifest="/tmp/cni/${CNI_MANIFESTS[$version]}"

# Wait for docker, else network might not be ready yet
while [[ `systemctl status docker | grep active | wc -l` -eq 0 ]]
do
    sleep 2
done

kubeadm init --config /etc/kubernetes/kubeadm.conf

# cni manifest is already configured at provision stage.
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f "$cni_manifest"

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

local_volume_manifest="/tmp/local-volume.yaml"
if [[ $minor_version -ge "16" ]]; then
    local_volume_manifest="/tmp/local-volume-ge-16.yaml"
fi
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f "$local_volume_manifest"
