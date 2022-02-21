#!/bin/bash

set -ex

if [ ! -f "/tmp/extra-pre-pull-images" ]; then
    echo "ERROR: extra-pre-pull-images list missing"
    exit 1
fi
if [ ! -f "/tmp/fetch-images.sh" ]; then
    echo "ERROR: fetch-images.sh missing"
    exit 1
fi

KUBEVIRTCI_SHARED_DIR=/var/lib/kubevirtci
mkdir -p $KUBEVIRTCI_SHARED_DIR
cat << EOF > $KUBEVIRTCI_SHARED_DIR/shared_vars.sh
#!/bin/bash
set -ex
export KUBELET_CGROUP_ARGS="--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice"
export KUBELET_FEATURE_GATES="IPv6DualStack=true"
export ISTIO_VERSION=1.13.0
export ISTIO_BIN_DIR=/opt/istio-$ISTIO_VERSION/bin
EOF
source $KUBEVIRTCI_SHARED_DIR/shared_vars.sh

function pull_container_retry() {
    retry=0
    maxRetries=5
    retryAfterSeconds=3
    until [ ${retry} -ge ${maxRetries} ]; do
        podman pull "$@" && break
        retry=$((${retry} + 1))
        echo "Retrying ${FUNCNAME[0]} [${retry}/${maxRetries}] in ${retryAfterSeconds}(s)"
        sleep ${retryAfterSeconds}
    done

    if [ ${retry} -ge ${maxRetries} ]; then
        echo "${FUNCNAME[0]} Failed after ${maxRetries} attempts!"
        exit 1
    fi
}

kubeadmn_patches_path="/provision/kubeadm-patches"

# Install modules of the initrd kernel
dnf install -y kernel-modules-$(uname -r)

# Resize root partition
dnf install -y cloud-utils-growpart
if growpart /dev/vda 1; then
    xfs_growfs -d /
fi

mkdir -p /provision

dnf install -y patch
cni_manifest="/provision/cni.yaml"
mv /tmp/cni.do-not-change.yaml $cni_manifest
patch $cni_manifest /tmp/cni.diff

cp /tmp/local-volume.yaml /provision/local-volume.yaml


systemctl stop firewalld || :
systemctl disable firewalld || :
# Make sure the firewall is never enabled again
# Enabling the firewall destroys the iptable rules
yum -y remove firewalld

# Required for iscsi demo to work.
yum -y install iscsi-initiator-utils

# for rook ceph
dnf -y install lvm2
# Convince ceph our storage is fast (not a rotational disk)
echo 'ACTION=="add|change", SUBSYSTEM=="block", KERNEL=="vd[a-z]", ATTR{queue/rotational}="0"' \
	> /etc/udev/rules.d/60-force-ssd-rotational.rules


# To prevent preflight issue related to tc not found
dnf install -y tc

# Install istioctl
export PATH=$ISTIO_BIN_DIR:$PATH
(
  set -E
  mkdir -p $ISTIO_BIN_DIR
  curl https://storage.googleapis.com/kubevirtci-istioctl-mirror/istio-$ISTIO_VERSION/bin/istioctl -o $ISTIO_BIN_DIR/istioctl
  chmod +x $ISTIO_BIN_DIR/istioctl
)
# generate Istio manifests for pre-pulling images
istioctl manifest generate --set profile=demo --set components.cni.enabled=true | tee /tmp/istio-deployment.yaml

export CRIO_VERSION=1.22
cat << EOF >/etc/yum.repos.d/devel_kubic_libcontainers_stable.repo
[devel_kubic_libcontainers_stable]
name=Stable Releases of Upstream github.com/containers packages (CentOS_8_Stream)
type=rpm-md
baseurl=https://storage.googleapis.com/kubevirtci-crio-mirror/devel_kubic_libcontainers_stable/
gpgcheck=0
enabled=1
EOF
cat << EOF >/etc/yum.repos.d/devel_kubic_libcontainers_stable_cri-o_${CRIO_VERSION}.repo
[devel_kubic_libcontainers_stable_cri-o_${CRIO_VERSION}]
name=devel:kubic:libcontainers:stable:cri-o:${CRIO_VERSION} (CentOS_8_Stream)
type=rpm-md
baseurl=https://storage.googleapis.com/kubevirtci-crio-mirror/devel_kubic_libcontainers_stable_cri-o_${CRIO_VERSION}
gpgcheck=0
enabled=1
EOF
dnf install -y cri-o

# install podman for functionality missing in crictl (tag, etc)
dnf install -y podman

# link docker to podman as we need docker in test repos to pre-pull images
# don't break them by doing a symlink
ln -s /usr/bin/podman /usr/bin/docker

cat << EOF > /etc/containers/registries.conf
[registries.search]
registries = ['registry.access.redhat.com', 'registry.fedoraproject.org', 'quay.io', 'docker.io']

[registries.insecure]
registries = ['registry:5000']

[registries.block]
registries = []
EOF

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
KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice  --fail-swap-on=false --kubelet-cgroups=/systemd/system.slice --feature-gates="IPv6DualStack=true"
EOT

# Needed for kubernetes service routing and dns
# https://github.com/kubernetes/kubernetes/issues/33798#issuecomment-250962627
modprobe bridge
modprobe overlay
modprobe br_netfilter
cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
net.ipv6.conf.all.disable_ipv6 = 0
net.ipv6.conf.all.forwarding = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF
sysctl --system

echo bridge >> /etc/modules-load.d/k8s.conf
echo br_netfilter >> /etc/modules-load.d/k8s.conf
echo overlay >> /etc/modules-load.d/k8s.conf

systemctl daemon-reload
systemctl enable crio && systemctl start crio
systemctl enable kubelet && systemctl start kubelet

dnf install -y NetworkManager

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

dnf install -y centos-release-nfv-openvswitch
dnf install -y openvswitch2.16

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
kubeadm init --config $kubeadm_manifest --ignore-preflight-errors=SWAP --experimental-patches /provision/kubeadm-patches/

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

# Pre pull all images from the manifests
for image in $(/tmp/fetch-images.sh /tmp); do
    pull_container_retry "${image}"
done

# Pre pull additional images from list
for image in $(cat "/tmp/extra-pre-pull-images"); do
    pull_container_retry "${image}"
done

# copy network addons operator manifests
# so we can use them at cluster-up
cp -rf /tmp/cnao/ /opt/

# copy whereabouts manifests
# so we can use them at cluster-up
cp -rf /tmp/whereabouts/ /opt/

# copy cdi manifests
cp -rf /tmp/cdi*.yaml /opt/

# Create a properly labelled tmp directory for testing
mkdir -p /var/provision/kubevirt.io/tests
chcon -t container_file_t /var/provision/kubevirt.io/tests
echo "tmpfs /var/provision/kubevirt.io/tests tmpfs rw,context=system_u:object_r:container_file_t:s0 0 1" >> /etc/fstab

dnf install -y NetworkManager-config-server

# Cleanup the existing NetworkManager profiles so the VM instances will come
# up with the default profiles. (Base VM image includes non default settings)
rm -f /etc/sysconfig/network-scripts/ifcfg-*
nmcli connection add con-name eth0 ifname eth0 type ethernet

# Remove machine-id, allowing unique ID/s for its instances
rm -f /etc/machine-id ; touch /etc/machine-id
