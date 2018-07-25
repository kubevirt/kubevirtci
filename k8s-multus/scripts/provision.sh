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

# Omit pgp checks until https://github.com/kubernetes/kubeadm/issues/643 is resolved.
yum install --nogpgcheck -y \
    kubeadm-${version} \
    kubelet-${version} \
    kubectl-${version} \
    kubernetes-cni \
    git \
    golang

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

kubeadm init --pod-network-cidr=10.244.0.0/16 --kubernetes-version v${version} --token abcdef.1234567890123456

# install multus
git clone https://github.com/intel/multus-cni.git -b dev/network-plumbing-working-group-crd-change
cd multus-cni
./build

cp bin/multus /opt/cni/bin/

cd examples

cat > macvlan-conf.yml <<EOF
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: macvlan-conf
spec: 
  config: '{
      "cniVersion": "0.3.0",
      "type": "macvlan",
      "master": "eth0",
      "mode": "bridge",
      "ipam": {
        "type": "host-local",
        "subnet": "192.168.66.0/24",
        "rangeStart": "192.168.66.200",
        "rangeEnd": "192.168.66.216",
        "routes": [
          { "dst": "0.0.0.0/0" }
        ],
        "gateway": "192.168.66.2"
      }
    }'
EOF

kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f clusterrole.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create clusterrolebinding multus-node-`hostname` \
                                                       --clusterrole=multus-crd-overpowered \
                                                       --user=system:node:`hostname`

curl https://raw.githubusercontent.com/intel/multus-cni/dev/network-plumbing-working-group-crd-change/examples/multus-with-flannel.yml --output /etc/kubernetes/cni.yml

kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /etc/kubernetes/cni.yml

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

cat > /etc/kubernetes/kubeadm.conf <<EOF
apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
apiServerExtraArgs:
  runtime-config: admissionregistration.k8s.io/v1alpha1
  ${admission_flag}: Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota
token: abcdef.1234567890123456
kubernetesVersion: ${version}
networking:
  podSubnet: 10.244.0.0/16
EOF