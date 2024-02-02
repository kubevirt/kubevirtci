#!/bin/bash

set -ex

source /var/lib/kubevirtci/shared_vars.sh

nodeip=
control_ip=192.168.66.101
if [ -f /home/vagrant/single_stack ]; then
    nodeip="--node-ip=::"
    control_ip=[fd00::101]
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

# Configure cgroup v2 settings
if [ -f /sys/fs/cgroup/cgroup.controllers ]; then
    echo "Configuring cgroup v2"

    CRIO_CONF_DIR=/etc/crio/crio.conf.d
    mkdir -p ${CRIO_CONF_DIR}
    cat << EOF > ${CRIO_CONF_DIR}/00-cgroupv2.conf
[crio.runtime]
conmon_cgroup = "pod"
cgroup_manager = "cgroupfs"
EOF

    sed -i 's/--cgroup-driver=systemd/--cgroup-driver=cgroupfs/' /var/lib/kubevirtci/shared_vars.sh
    source /var/lib/kubevirtci/shared_vars.sh

    # kubelet will be started later on
    systemctl stop kubelet
    systemctl restart crio
fi

# Wait for crio, else network might not be ready yet
while [[ `systemctl status crio | grep active | wc -l` -eq 0 ]]
do
    sleep 2
done

if [ -f /etc/sysconfig/kubelet ]; then
    # TODO use config file! this is deprecated
    cat <<EOT >>/etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=${KUBELET_CGROUP_ARGS} --fail-swap-on=false ${nodeip} --feature-gates=CPUManager=true,NodeSwap=true --cpu-manager-policy=static --kube-reserved=cpu=250m --system-reserved=cpu=250m
EOT
else
    cat <<EOT >>/etc/systemd/system/kubelet.service.d/09-kubeadm.conf
Environment="KUBELET_CPUMANAGER_ARGS=--fail-swap-on=false --feature-gates=CPUManager=true,NodeSwap=true ${nodeip} --cpu-manager-policy=static --kube-reserved=cpu=250m --system-reserved=cpu=250m"
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

# Disable swap
sudo swapoff -a

until ip address show dev eth0 | grep global | grep inet6; do sleep 1; done

kubeadm join --token abcdef.1234567890123456 ${control_ip}:6443 --ignore-preflight-errors=all --discovery-token-unsafe-skip-ca-verification=true

# ceph mon permission
mkdir -p /var/lib/rook
chcon -t container_file_t /var/lib/rook
