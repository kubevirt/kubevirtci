#!/bin/bash

set -ex

setenforce 0
sed -i "s/^SELINUX=.*/SELINUX=permissive/" /etc/selinux/config

# Disable swap
swapoff -a
sed -i '/ swap / s/^/#/' /etc/fstab

# Disable spectre and meltdown patches
sed -i 's/quiet"/quiet spectre_v2=off nopti hugepagesz=2M hugepages=64"/' /etc/default/grub
grub2-mkconfig -o /boot/grub2/grub.cfg

systemctl stop firewalld NetworkManager || :
systemctl disable firewalld NetworkManager || :
# Make sure the firewall is never enabled again
# Enabling the firewall destroys the iptable rules
yum -y remove NetworkManager firewalld

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
yum install -y docker

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
    kubernetes-cni

# Latest docker on CentOS uses systemd for cgroup management
# kubeadm 1.11 uses a new config method for the kubelet
if [[ $version =~ \.([0-9]+) ]] && [[ ${BASH_REMATCH[1]} -ge "11" ]]; then
    # TODO use config file! this is deprecated
    cat <<EOT >/etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice --feature-gates=BlockVolume=true
EOT
else
    cat <<EOT >>/etc/systemd/system/kubelet.service.d/09-kubeadm.conf
[Service]
Environment="KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice --feature-gates=BlockVolume=true"
EOT
fi

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

kubeadm init --pod-network-cidr=10.244.0.0/16 --kubernetes-version v${version} --token abcdef.1234567890123456
kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/coreos/flannel/v0.9.1/Documentation/kube-flannel.yml

# Wait at least for one pod
while [ -z "$(kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system | grep kube)" ]; do
    echo "Waiting for at least one pod ..."
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
# k8s 1.11 needs some changes
if [[ $version =~ \.([0-9]+) ]] && [[ ${BASH_REMATCH[1]} -ge "11" ]]; then
    # k8s 1.11 asks for confirmation on kubeadm reset, which can be suppressed by a new force flag
    reset_command="kubeadm reset --force"

    # k8s 1.11 uses new flags for admission plugins
    # old one is deprecated only, but can not be combined with new one, which is used in api server config created by kubeadm
    admission_flag="enable-admission-plugins"
fi

$reset_command

# TODO new format since 1.11, this old format will be removed with 1.12, see https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm-init/#config-file
cat > /etc/kubernetes/kubeadm.conf <<EOF
apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
apiServerExtraArgs:
  runtime-config: admissionregistration.k8s.io/v1alpha1
  ${admission_flag}: Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,CustomResourceSubresources
  feature-gates: "BlockVolume=true"
controllerManagerExtraArgs:
  feature-gates: "BlockVolume=true"
token: abcdef.1234567890123456
kubernetesVersion: ${version}
networking:
  podSubnet: 10.244.0.0/16
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
