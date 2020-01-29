#!/bin/bash

set -ex

# Wait for docker, else network might not be ready yet
while [[ `systemctl status docker | grep active | wc -l` -eq 0 ]]
do
    sleep 2
done

kubeadm init --config /etc/kubernetes/kubeadm.conf

default_cidr="192.168.0.0/16"
pod_cidr="10.244.0.0/16"
version=`kubectl version --short --client | cut -d":" -f2 |sed  's/ //g' | cut -c2- | cut -d"." -f2`

network_plugin_manifest="/tmp/flannel.yaml"
if [[ $version -ge "17" ]]; then
    network_plugin_manifest="/tmp/calico.yaml"
    sed -i -e "s?$default_cidr?$pod_cidr?g" "$network_plugin_manifest"
elif [[ $version -ge "16" ]]; then
    network_plugin_manifest="/tmp/flannel-ge-16.yaml"
elif [[ $version -ge "12" ]]; then
    network_plugin_manifest="/tmp/flannel-ge-12.yaml"
fi
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f "$network_plugin_manifest"

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
if [[ $version -ge "16" ]]; then
    local_volume_manifest="/tmp/local-volume-ge-16.yaml"
fi
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f "$local_volume_manifest"
