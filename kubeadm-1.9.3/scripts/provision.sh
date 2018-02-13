#!/bin/bash

set -ex

setenforce 0
sed -i "s/^SELINUX=.*/SELINUX=permissive/" /etc/selinux/config

# Disable swap
swapoff -a
sed -i '/ swap / s/^/#/' /etc/fstab

systemctl stop firewalld NetworkManager || :
systemctl disable firewalld NetworkManager || :
# Make sure the firewall is never enabled again
# Enabling the firewall destroys the iptable rules
yum -y remove NetworkManager firewalld

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

# Omit pgp checks until https://github.com/kubernetes/kubeadm/issues/643 is resolved.
yum install --nogpgcheck -y \
    kubeadm \
    kubelet \
    kubectl \
    kubernetes-cni

# Latest docker on CentOS uses systemd for cgroup management
cat <<EOT >>/etc/systemd/system/kubelet.service.d/09-kubeadm.conf
[Service]
Environment="KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice"
EOT
systemctl daemon-reload

systemctl enable docker && systemctl start docker
systemctl enable kubelet && systemctl start kubelet

# Needed for kubernetes service routing and dns
# https://github.com/kubernetes/kubernetes/issues/33798#issuecomment-250962627
modprobe bridge
cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sysctl --system

kubeadm init --pod-network-cidr=10.244.0.0/16 --kubernetes-version v1.9.3 --token abcdef.1234567890123456
kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/coreos/flannel/v0.9.1/Documentation/kube-flannel.yml
kubeadm reset
