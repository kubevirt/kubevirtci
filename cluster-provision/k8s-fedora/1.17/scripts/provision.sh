#!/bin/bash

set -ex

source /tmp/scripts/cnis-map.sh
source /tmp/scripts/config-cni.sh

function get_minor_version() {
    [[ $1 =~ \.([0-9]+) ]]
    echo ${BASH_REMATCH[1]}
}

cni_manifest="/tmp/cni/${CNI_MANIFESTS[$version]}"

minor_version=$(get_minor_version $version)

# Kubernetes rqeuierment: Set SELinux in permissive mode (effectively disabling it)
setenforce 0
sed -i "s/^SELINUX=.*/SELINUX=permissive/" /etc/selinux/config

# Kubernetes rqeuierment: Disable swap
swapoff -a
sed -i '/ swap / s/^/#/' /etc/fstab

systemctl stop firewalld || :
systemctl disable firewalld || :
# Make sure the firewall is never enabled again
# Enabling the firewall destroys the iptable rules
yum -y remove firewalld

# Required for iscsi demo to work.
yum -y install iscsi-initiator-utils

# Install docker required packages.
yum -y install yum-utils \
    device-mapper-persistent-data \
    lvm2

# Add Docker repository.
yum-config-manager --add-repo \
  https://download.docker.com/linux/centos/docker-ce.repo

# Install Docker CE.
yum -y install \
  containerd.io-1.2.10 \
  docker-ce-19.03.4 \
  docker-ce-cli-19.03.4

# Create /etc/docker directory.
mkdir /etc/docker

# Setup docker daemon
cat << EOF > /etc/docker/daemon.json
{
  "insecure-registries" : ["registry:5000"],
  "log-driver": "json-file",
  "exec-opts": ["native.cgroupdriver=systemd"]
}
EOF

mkdir -p /etc/systemd/system/docker.service.d

# Restart Docker
systemctl daemon-reload
systemctl restart docker

# Add Kubernetes repository.
cat << EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF

setenforce 0
sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config

# Install Kubernetes packages.
yum install --nogpgcheck --disableexcludes=kubernetes -y \
    kubeadm-${version} \
    kubelet-${version} \
    kubectl-${version} \
    kubernetes-cni-0.6.0

# Ensure iptables tooling does not use the nftables backend
update-alternatives --set iptables /usr/sbin/iptables-legacy

# Enable and state kubelet service
systemctl enable --now kubelet


if [[ $minor_version -ge "15" ]]; then
    # TODO use config file! this is deprecated
    cat <<EOT >/etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice --feature-gates="BlockVolume=true,CSIBlockVolume=true,VolumeSnapshotDataSource=true"
EOT
elif [[ $minor_version -ge "12" ]]; then
    # TODO use config file! this is deprecated
    cat <<EOT >/etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice --allow-privileged=true --feature-gates="BlockVolume=true,CSIBlockVolume=true,VolumeSnapshotDataSource=true"
EOT
elif [[ $minor_version -ge "11" ]]; then
    # TODO use config file! this is deprecated
    cat <<EOT >/etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice --allow-privileged=true --feature-gates="BlockVolume=true,CSIBlockVolume=true"
EOT
else
    cat <<EOT >>/etc/systemd/system/kubelet.service.d/09-kubeadm.conf
[Service]
Environment="KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice --allow-privileged=true --feature-gates=BlockVolume=true"
EOT
fi

# Start and enable docker and kubelet services, before instllation
systemctl daemon-reload

systemctl enable docker && systemctl start docker
systemctl enable kubelet && systemctl start kubelet

# configure additional settings for cni plugin
configure_cni $cni_manifest

kubeadm init --pod-network-cidr=$pod_cidr --kubernetes-version v${version} --token abcdef.1234567890123456

# Install CNI plugin
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

# Print cluster pods.
kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system

# Remove this node from cluster
reset_command="kubeadm reset"
admission_flag="admission-control"
# k8s 1.11 needs some changes
if [[ $minor_version -ge "11" ]]; then
    # k8s 1.11 asks for confirmation on kubeadm reset, which can be suppressed by a new force flag
    reset_command="kubeadm reset --force"

    # k8s 1.11 uses new flags for admission plugins
    # old one is deprecated only, but can not be combined with new one, which is used in api server config created by kubeadm
    admission_flag="enable-admission-plugins"
fi

# Remove kubernetes cluster
$reset_command

# audit log configuration
mkdir /etc/kubernetes/audit

# Export cluster policy
if [[ $minor_version -ge "12" ]]; then
    audit_api_version="audit.k8s.io/v1"
else
    audit_api_version="audit.k8s.io/v1beta1"
fi
if [[ $minor_version -ge "11" ]]; then
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
fi

# Export kubeadm configurations
if [[ $minor_version -ge "14" ]]; then
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

elif [[ $minor_version -ge "12" ]]; then
    cat > /etc/kubernetes/kubeadm.conf <<EOF
apiVersion: kubeadm.k8s.io/v1alpha3
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
apiServerExtraArgs:
  enable-admission-plugins: Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota
  feature-gates: "BlockVolume=true,CSIBlockVolume=true,VolumeSnapshotDataSource=true,AdvancedAuditing=true"
  allow-privileged: "true"
  runtime-config: admissionregistration.k8s.io/v1alpha1
  audit-policy-file: "/etc/kubernetes/audit/adv-audit.yaml"
  audit-log-path: "/var/log/k8s-audit/k8s-audit.log"
  audit-log-format: "json"
apiServerExtraVolumes:
- name: audit-conf
  hostPath: "/etc/kubernetes/audit"
  mountPath: "/etc/kubernetes/audit"
- name: audit-log
  hostPath: "/var/log/k8s-audit"
  mountPath: "/var/log/k8s-audit"
  writable: true
apiVersion: kubeadm.k8s.io/v1alpha3
controllerManagerExtraArgs:
  feature-gates: "BlockVolume=true,CSIBlockVolume=true,VolumeSnapshotDataSource=true"
kind: ClusterConfiguration
kubernetesVersion: ${version}
networking:
  podSubnet: 10.244.0.0/16

EOF

elif [[ $minor_version -ge "11" ]]; then
    cat > /etc/kubernetes/kubeadm.conf <<EOF
apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
apiServerExtraArgs:
  runtime-config: admissionregistration.k8s.io/v1alpha1
  ${admission_flag}: Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota
  feature-gates: "BlockVolume=true,CustomResourceSubresources=true,CSIBlockVolume=true,AdvancedAuditing=true"
  allow-privileged: "true"
  audit-policy-file: "/etc/kubernetes/audit/adv-audit.yaml"
  audit-log-path: "/var/log/k8s-audit/k8s-audit.log"
  audit-log-format: "json"
apiServerExtraVolumes:
- name: audit-conf
  hostPath: "/etc/kubernetes/audit"
  mountPath: "/etc/kubernetes/audit"
- name: audit-log
  hostPath: "/var/log/k8s-audit"
  mountPath: "/var/log/k8s-audit"
  writable: true
controllerManagerExtraArgs:
  feature-gates: "BlockVolume=true,CSIBlockVolume=true"
token: abcdef.1234567890123456
kubernetesVersion: ${version}
networking:
  podSubnet: 10.244.0.0/16
EOF

else
    cat > /etc/kubernetes/kubeadm.conf <<EOF
apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
apiServerExtraArgs:
  runtime-config: admissionregistration.k8s.io/v1alpha1
  ${admission_flag}: Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota
  feature-gates: "BlockVolume=true,CustomResourceSubresources=true"
  allow-privileged: "true"
controllerManagerExtraArgs:
  feature-gates: "BlockVolume=true"
token: abcdef.1234567890123456
kubernetesVersion: ${version}
networking:
  podSubnet: 10.244.0.0/16
EOF
fi

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
