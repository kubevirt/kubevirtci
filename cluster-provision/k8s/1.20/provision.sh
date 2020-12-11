#!/bin/bash

set -ex

function docker_pull_retry() {
    retry=0
    maxRetries=5
    retryAfterSeconds=3
    until [ ${retry} -ge ${maxRetries} ]; do
        docker pull $@ && break
        retry=$((${retry} + 1))
        echo "Retrying ${FUNCNAME} [${retry}/${maxRetries}] in ${retryAfterSeconds}(s)"
        sleep ${retryAfterSeconds}
    done

    if [ ${retry} -ge ${maxRetries} ]; then
        echo "${FUNCNAME} Failed after ${maxRetries} attempts!"
        exit 1
    fi
}

kubeadmn_patches_path="/provision/kubeadm-patches"

# Need to have the latest kernel
dnf update -y kernel

# Resize root partition
dnf install -y cloud-utils-growpart
if growpart /dev/vda 1; then
    xfs_growfs -d /
fi

mkdir -p /provision

dnf install -y patch
cni_manifest="/provision/cni.yaml"
cp /tmp/cni.do-not-change.yaml $cni_manifest
patch $cni_manifest /tmp/cni.diff

cp /tmp/local-volume.yaml /provision/local-volume.yaml

# Disable swap
swapoff -a
sed -i '/ swap / s/^/#/' /etc/fstab

# Disable spectre and meltdown patches
echo 'GRUB_CMDLINE_LINUX="${GRUB_CMDLINE_LINUX} spectre_v2=off nopti hugepagesz=2M hugepages=64 intel_iommu=on modprobe.blacklist=nouveau"' >> /etc/default/grub
grub2-mkconfig -o /boot/grub2/grub.cfg

systemctl stop firewalld || :
systemctl disable firewalld || :
# Make sure the firewall is never enabled again
# Enabling the firewall destroys the iptable rules
yum -y remove firewalld

# Required for iscsi demo to work.
yum -y install iscsi-initiator-utils

# To prevent preflight issue realted to tc not found
dnf install -y tc

# Install docker required packages.
dnf -y install yum-utils \
    device-mapper-persistent-data \
    lvm2

# Add Docker repository.
dnf config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo

# Enable the nightly version to get a new enough runc version which includes containerd
# with a new enough runc version to avoid https://github.com/kubernetes/kubernetes/issues/95296#issuecomment-714419461
# which manifests in /dev/* permission denied flakes on cpu manager enabled nodes
dnf config-manager --set-enabled docker-ce-nightly

# Install package(s) that trigger the enablement of the container-tools yum module
dnf -y install container-selinux

# Disable the container-tools module, as it forces containerd.io to an ancient version,
#   which in turn forces docker-ce to an older version, making it incompatible with docker-ce-cli...
dnf -y module disable container-tools

# Install iptables needed for docker-ce
dnf install -y iptables

# Install Docker CE.
dnf install -y docker-ce --nobest

# Create /etc/docker directory.
mkdir /etc/docker

# Setup docker daemon
cat << EOF > /etc/docker/daemon.json
{
  "insecure-registries" : ["registry:5000"],
  "log-driver": "json-file",
  "exec-opts": ["native.cgroupdriver=systemd"],
  "ipv6": true,
  "fixed-cidr-v6": "2001:db9:1::/64",
  "selinux-enabled": true
}
EOF

mkdir -p /etc/systemd/system/docker.service.d

# Restart Docker
systemctl daemon-reload
systemctl restart docker

#TODO: el8 repo
# Add Kubernetes repository.
cat <<EOF >/etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF

# Install Kubernetes packages.
dnf install --skip-broken --nobest --nogpgcheck --disableexcludes=kubernetes -y \
    kubeadm-${version} \
    kubelet-${version} \
    kubectl-${version} \
    kubernetes-cni

# TODO use config file! this is deprecated
cat <<EOT >/etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice --feature-gates="BlockVolume=true,CSIBlockVolume=true,VolumeSnapshotDataSource=true,IPv6DualStack=true"
EOT

systemctl daemon-reload

systemctl enable docker && systemctl start docker
systemctl enable kubelet && systemctl start kubelet

# Needed for kubernetes service routing and dns
# https://github.com/kubernetes/kubernetes/issues/33798#issuecomment-250962627
modprobe bridge
modprobe br_netfilter
cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
net.ipv6.conf.all.disable_ipv6 = 0
net.ipv6.conf.all.forwarding = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF
sysctl --system

echo bridge >> /etc/modules
echo br_netfilter >> /etc/modules

NM_VERSION=1.30.0
dnf install -y NetworkManager-$NM_VERSION

# configure additional settings for cni plugin
cat <<EOF >/etc/NetworkManager/conf.d/001-calico.conf
[keyfile]
unmanaged-devices=interface-name:cali*;interface-name:tunl*
EOF

# Use dhclient to have expected hostname behaviour
cat <<EOF >/etc/NetworkManager/conf.d/002-dhclient.conf
[main]
dhcp=dhclient
EOF

sysctl -w net.netfilter.nf_conntrack_max=1000000
echo "net.netfilter.nf_conntrack_max=1000000" >> /etc/sysctl.conf

systemctl restart NetworkManager

nmcli connection modify "System eth0" \
   ipv6.method auto \
   ipv6.addr-gen-mode eui64
nmcli connection up "System eth0"

mkdir -p $kubeadmn_patches_path

cat >$kubeadmn_patches_path/kube-apiserver.yaml <<EOF
spec:
  securityContext:
    seLinuxOptions:
      type: spc_t
EOF
cat >$kubeadmn_patches_path/kube-controller-manager.yaml <<EOF
spec:
  securityContext:
    seLinuxOptions:
      type: spc_t
EOF
cat >$kubeadmn_patches_path/kube-scheduler.yaml <<EOF
spec:
  securityContext:
    seLinuxOptions:
      type: spc_t
EOF
cat >$kubeadmn_patches_path/etcd.yaml <<EOF
spec:
  securityContext:
    seLinuxOptions:
      type: spc_t
EOF

cat >$kubeadmn_patches_path/add-security-context-deployment-patch.yaml <<EOF
spec:
  template:
    spec:
      securityContext:
        seLinuxOptions:
          type: spc_t
EOF

# audit log configuration
mkdir /etc/kubernetes/audit

audit_api_version="audit.k8s.io/v1"
cat > /etc/kubernetes/audit/adv-audit.yaml <<EOF
apiVersion: ${audit_api_version}
kind: Policy
rules:
- level: Request
  users: ["kubernetes-admin"]
  resources:
  - group: kubevirt.io
    resources:
    - virtualmachines
    - virtualmachineinstances
    - virtualmachineinstancereplicasets
    - virtualmachineinstancepresets
    - virtualmachineinstancemigrations
  omitStages:
  - RequestReceived
  - ResponseStarted
  - Panic
EOF

kubeadm_manifest="/etc/kubernetes/kubeadm.conf"
envsubst < /tmp/kubeadm.conf > $kubeadm_manifest
kubeadm init --config $kubeadm_manifest --experimental-patches /provision/kubeadm-patches/

kubectl --kubeconfig=/etc/kubernetes/admin.conf patch deployment coredns -n kube-system -p "$(cat $kubeadmn_patches_path/add-security-context-deployment-patch.yaml)"
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f "$cni_manifest"

# Wait at least for 7 pods
while [[ "$(kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system --no-headers | wc -l)" -lt 7 ]]; do
    echo "Waiting for at least 7 pods to appear ..."
    kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system
    sleep 10
done

# Wait until k8s pods are running
while [ -n "$(kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system --no-headers | grep -v Running)" ]; do
    echo "Waiting for k8s pods to enter the Running state ..."
    kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system --no-headers | >&2 grep -v Running || true
    sleep 10
done

# Make sure all containers are ready
while [ -n "$(kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system -o'custom-columns=status:status.containerStatuses[*].ready,metadata:metadata.name' --no-headers | grep false)" ]; do
    echo "Waiting for all containers to become ready ..."
    kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system -o'custom-columns=status:status.containerStatuses[*].ready,metadata:metadata.name' --no-headers
    sleep 10
done

kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system

kubeadm reset --force

# Create local-volume directories
for i in {1..10}
do
  mkdir -p /var/local/kubevirt-storage/local-volume/disk${i}
  mkdir -p /mnt/local-storage/local/disk${i}
  echo "/var/local/kubevirt-storage/local-volume/disk${i} /mnt/local-storage/local/disk${i} none defaults,bind 0 0" >> /etc/fstab
done
chmod -R 777 /var/local/kubevirt-storage/local-volume

# Setup selinux permissions to local volume directories.
chcon -R unconfined_u:object_r:svirt_sandbox_file_t:s0 /mnt/local-storage/

# Pre pull fluentd image used in logging
docker_pull_retry fluent/fluentd:v1.2-debian
docker_pull_retry fluent/fluentd-kubernetes-daemonset:v1.2-debian-syslog

# Pre pull images used in Ceph CSI
docker_pull_retry quay.io/k8scsi/csi-attacher:v1.0.1
docker_pull_retry quay.io/k8scsi/csi-provisioner:v1.0.1
docker_pull_retry quay.io/k8scsi/csi-snapshotter:v1.0.1
docker_pull_retry quay.io/cephcsi/rbdplugin:v1.0.0
docker_pull_retry quay.io/k8scsi/csi-node-driver-registrar:v1.0.2

# Pre pull cluster network addons operator images and store manifests
# so we can use them at cluster-up
cp -rf /tmp/cnao/ /opt/
for i in $(grep -A 2 "IMAGE" /opt/cnao/operator.yaml | grep value | awk '{print $2}'); do docker_pull_retry $i; done

# Pre pull local-volume-provisioner
for i in $(grep -A 2 "IMAGE" /provision/local-volume.yaml | grep value | awk -F\" '{print $2}'); do docker_pull_retry $i; done

# Create a properly labelled tmp directory for testing
mkdir -p /var/provision/kubevirt.io/tests
chcon -t container_file_t /var/provision/kubevirt.io/tests
echo "tmpfs /var/provision/kubevirt.io/tests tmpfs rw,context=system_u:object_r:container_file_t:s0 0 1" >> /etc/fstab

dnf install -y NetworkManager-config-server-$NM_VERSION

# Cleanup the existing NetworkManager profiles so the VM instances will come
# up with the default profiles. (Base VM image includes non default settings)
rm -f /etc/sysconfig/network-scripts/ifcfg-*
nmcli connection add con-name eth0 ifname eth0 type ethernet

# Remove machine-id, allowing unique ID/s for its instances
rm -f /etc/machine-id ; touch /etc/machine-id
