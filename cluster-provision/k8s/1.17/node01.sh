#!/bin/bash

set -ex

# Ensure that hugepages are there
# Hugetlb holds total huge page size in kB including both 2M or 1G hugepages
HUGETLB=`cat /proc/meminfo | sed -e "s/ //g" | grep "Hugetlb:"`
HUGEPAGE=(${HUGETLB//:/ })
HUGEPAGE_PARTS=(${HUGEPAGE[-1]//kB/ })
HUGEPAGE_TOTAL=${HUGEPAGE_PARTS[0]}
if [[ $HUGEPAGE_TOTAL -lt $((64 * 2048)) ]]; then
    echo "Minimum of 64 2M Hugepages is required"
    exit 1
fi

timeout=30
interval=5
while ! hostnamectl  |grep Transient ; do
    echo "Waiting for dhclient to set the hostname from dnsmasq"
    sleep $interval
    timeout=$(( $timeout - $interval ))
    if [ $timeout -le 0 ]; then
        exit 1
    fi
done

version=`kubectl version --short --client | cut -d":" -f2 |sed  's/ //g' | cut -c2- `
cni_manifest="/var/provision/cni.yaml"

# Wait for docker, else network might not be ready yet
while [[ `systemctl status docker | grep active | wc -l` -eq 0 ]]
do
    sleep 2
done

kubeadm init --config /etc/kubernetes/kubeadm.conf --experimental-kustomize /var/provision/kubeadm-patches/

kubectl --kubeconfig=/etc/kubernetes/admin.conf patch deployment coredns -n kube-system -p "$(cat /var/provision/kubeadm-patches/add-security-context-deployment-patch.yaml)"
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

local_volume_manifest="/var/provision/local-volume.yaml"
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f "$local_volume_manifest"
