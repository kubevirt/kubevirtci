#!/bin/bash

set -ex

cni_manifest="/tmp/cni.yaml"

setenforce 0
sed -i "s/^SELINUX=.*/SELINUX=permissive/" /etc/selinux/config

# Disable swap
swapoff -a
sed -i '/ swap / s/^/#/' /etc/fstab

# Disable spectre and meltdown patches
echo 'GRUB_CMDLINE_LINUX="${GRUB_CMDLINE_LINUX} spectre_v2=off nopti hugepagesz=2M hugepages=64"' >> /etc/default/grub
grub2-mkconfig -o /boot/grub2/grub.cfg

systemctl stop firewalld || :
systemctl disable firewalld || :
# Make sure the firewall is never enabled again
# Enabling the firewall destroys the iptable rules
yum -y remove firewalld

# Required for iscsi demo to work.
yum -y install iscsi-initiator-utils

cat <<EOF >/etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=http://yum.kubernetes.io/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg
       https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF
yum install -y \
  docker-1.13.1 \
  python-docker-py-1.10.6 \
  python3-pip

pip3 install --default-timeout=100 docker-pycreds

# Log to json files instead of journald
sed -i 's/--log-driver=journald //g' /etc/sysconfig/docker
echo '{ "insecure-registries" : ["registry:5000"] }' > /etc/docker/daemon.json

# Enable the permanent logging
# Required by the fluentd journald plugin
# The default settings in recent distribution for systemd is set to auto,
# when on auto journal is permament when /var/log/journal exists
mkdir -p /var/log/journal

# Omit pgp checks until https://github.com/kubernetes/kubeadm/issues/643 is resolved.
yum install --nogpgcheck -y \
    kubeadm-${version} \
    kubelet-${version} \
    kubectl-${version} \
    kubernetes-cni-0.6.0

# TODO use config file! this is deprecated
cat <<EOT >/etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice --allow-privileged=true --feature-gates="BlockVolume=true,CSIBlockVolume=true,VolumeSnapshotDataSource=true"
EOT

systemctl daemon-reload

systemctl enable docker && systemctl start docker
systemctl enable kubelet && systemctl start kubelet

# Needed for kubernetes service routing and dns
# https://github.com/kubernetes/kubernetes/issues/33798#issuecomment-250962627
modprobe bridge
modprobe br_netfilter
cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
EOF
sysctl --system

echo bridge >> /etc/modules
echo br_netfilter >> /etc/modules

default_cidr="192.168.0.0/16"
pod_cidr="10.244.0.0/16"
kubeadm init --pod-network-cidr=$pod_cidr --kubernetes-version v${version} --token abcdef.1234567890123456

sed -i -e "s?$default_cidr?$pod_cidr?g" $cni_manifest 
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

reset_command="kubeadm reset"
admission_flag="admission-control"
# k8s 1.11 asks for confirmation on kubeadm reset, which can be suppressed by a new force flag
reset_command="kubeadm reset --force"

# k8s 1.11 uses new flags for admission plugins
# old one is deprecated only, but can not be combined with new one, which is used in api server config created by kubeadm
admission_flag="enable-admission-plugins"

$reset_command

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

cat > /etc/kubernetes/kubeadm.conf <<EOF
apiVersion: kubeadm.k8s.io/v1beta1
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token
  token: abcdef.1234567890123456
  ttl: 24h0m0s
  usages:
  - signing
  - authentication
kind: InitConfiguration
---
apiServer:
  extraArgs:
    allow-privileged: "true"
    audit-log-format: json
    audit-log-path: /var/log/k8s-audit/k8s-audit.log
    audit-policy-file: /etc/kubernetes/audit/adv-audit.yaml
    enable-admission-plugins: NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota
    feature-gates: BlockVolume=true,CSIBlockVolume=true,VolumeSnapshotDataSource=true,AdvancedAuditing=true
  extraVolumes:
  - hostPath: /etc/kubernetes/audit
    mountPath: /etc/kubernetes/audit
    name: audit-conf
    readOnly: true
  - hostPath: /var/log/k8s-audit
    mountPath: /var/log/k8s-audit
    name: audit-log
  timeoutForControlPlane: 4m0s
apiVersion: kubeadm.k8s.io/v1beta1
certificatesDir: /etc/kubernetes/pki
clusterName: kubernetes
controlPlaneEndpoint: ""
controllerManager:
  extraArgs:
    feature-gates: BlockVolume=true,CSIBlockVolume=true,VolumeSnapshotDataSource=true
dns:
  type: CoreDNS
etcd:
  local:
    dataDir: /var/lib/etcd
imageRepository: k8s.gcr.io
kind: ClusterConfiguration
kubernetesVersion: ${version}
networking:
  dnsDomain: cluster.local
  podSubnet: 10.244.0.0/16
  serviceSubnet: 10.96.0.0/12
EOF

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
docker pull fluent/fluentd:v1.2-debian
docker pull fluent/fluentd-kubernetes-daemonset:v1.2-debian-syslog

# Pre pull images used in Ceph CSI
docker pull quay.io/k8scsi/csi-attacher:v1.0.1
docker pull quay.io/k8scsi/csi-provisioner:v1.0.1
docker pull quay.io/k8scsi/csi-snapshotter:v1.0.1
docker pull quay.io/cephcsi/rbdplugin:v1.0.0
docker pull quay.io/k8scsi/csi-node-driver-registrar:v1.0.2

# Create a properly labelled tmp directory for testing
mkdir -p /tmp/kubevirt.io/tests
chcon -t container_file_t /tmp/kubevirt.io/tests
echo "tmpfs /tmp/kubevirt.io/tests tmpfs rw,context=system_u:object_r:container_file_t:s0 0 1" >> /etc/fstab
