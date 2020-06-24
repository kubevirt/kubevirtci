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

# Wait for docker, else network might not be ready yet
while [[ `systemctl status docker | grep active | wc -l` -eq 0 ]]
do
    sleep 2
done

if [ -f /etc/sysconfig/kubelet ]; then
    # TODO use config file! this is deprecated
    cat <<EOT >>/etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=${KUBELET_EXTRA_ARGS} --feature-gates=CPUManager=true --cpu-manager-policy=static --kube-reserved=cpu=500m --system-reserved=cpu=500m
EOT
else
    cat <<EOT >>/etc/systemd/system/kubelet.service.d/09-kubeadm.conf
Environment="KUBELET_CPUMANAGER_ARGS=--feature-gates=CPUManager=true --cpu-manager-policy=static --kube-reserved=cpu=500m --system-reserved=cpu=500m"
EOT
sed -i 's/$KUBELET_EXTRA_ARGS/$KUBELET_EXTRA_ARGS $KUBELET_CPUMANAGER_ARGS/' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
fi

systemctl daemon-reload
service kubelet restart
kubelet_rc=$?
if [[ $kubelet_rc -ne 0 ]]; then
    rm -rf /var/lib/kubelet/cpu_manager_state
    service kubelet restart
fi

kubeadm join --token abcdef.1234567890123456 192.168.66.101:6443 --ignore-preflight-errors=all --discovery-token-unsafe-skip-ca-verification=true
