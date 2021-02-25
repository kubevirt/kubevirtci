#!/bin/bash

set -ex

source /var/lib/kubevirtci/kubelet_args.sh

# Ensure that hugepages are there
cat /proc/meminfo | sed -e "s/ //g" | grep "HugePages_Total:64"

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

# Wait for crio, else network might not be ready yet
while [[ `systemctl status crio | grep active | wc -l` -eq 0 ]]
do
    sleep 2
done

if [ -f /etc/sysconfig/kubelet ]; then
    # TODO use config file! this is deprecated
    cat <<EOT >>/etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=${KUBELET_CGROUP_ARGS} --feature-gates=${KUBELET_FEATURE_GATES},CPUManager=true --cpu-manager-policy=static --kube-reserved=cpu=500m --system-reserved=cpu=500m
EOT
else
    cat <<EOT >>/etc/systemd/system/kubelet.service.d/09-kubeadm.conf
Environment="KUBELET_CPUMANAGER_ARGS=--feature-gates=CPUManager=true,IPv6DualStack=true --cpu-manager-policy=static --kube-reserved=cpu=500m --system-reserved=cpu=500m"
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
