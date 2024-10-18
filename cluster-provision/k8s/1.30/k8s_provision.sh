#!/bin/bash

set -ex

source /var/lib/kubevirtci/shared_vars.sh

function getKubernetesClosestStableVersion() {
  kubernetes_version=$version
  packages_version=$kubernetes_version
  if [[ $kubernetes_version == *"alpha"* ]] || [[ $kubernetes_version == *"beta"* ]] || [[ $kubernetes_version == *"rc"* ]]; then
    kubernetes_minor_version=$(echo $kubernetes_version | cut -d. -f2)
    packages_major_version=$(echo $kubernetes_version | cut -d. -f1)
    packages_minor_version=$((kubernetes_minor_version-1))
    packages_version="$(curl --fail -L "https://storage.googleapis.com/kubernetes-release/release/stable-${packages_major_version}.${packages_minor_version}.txt" | sed 's/^v//')"
  fi
  echo $packages_version
}

function replaceKubeBinaries() {
  dnf install -y which
  rm -f $(which kubeadm kubelet)

  BIN_DIR="/usr/bin"
  RELEASE="v$version"
  ARCH="amd64"

  curl -L --remote-name-all https://dl.k8s.io/release/${RELEASE}/bin/linux/${ARCH}/kubeadm -o ${BIN_DIR}/kubeadm
  curl -L --remote-name-all https://dl.k8s.io/release/${RELEASE}/bin/linux/${ARCH}/kubelet -o ${BIN_DIR}/kubelet
  chmod +x ${BIN_DIR}/kubeadm ${BIN_DIR}/kubelet
}

if [ ! -f "/tmp/extra-pre-pull-images" ]; then
    echo "ERROR: extra-pre-pull-images list missing"
    exit 1
fi
if [ ! -f "/tmp/fetch-images.sh" ]; then
    echo "ERROR: fetch-images.sh missing"
    exit 1
fi

if grep -q "CentOS Stream 9" /etc/os-release; then
  release="centos9"
else
  echo "ERROR: Could not recognize guest OS"
  exit 1
fi

function pull_container_retry() {
    retry=0
    maxRetries=5
    retryAfterSeconds=3
    until [ ${retry} -ge ${maxRetries} ]; do
        crictl pull "$@" && break
        retry=$((${retry} + 1))
        echo "Retrying ${FUNCNAME[0]} [${retry}/${maxRetries}] in ${retryAfterSeconds}(s)"
        sleep ${retryAfterSeconds}
    done

    if [ ${retry} -ge ${maxRetries} ]; then
        echo "${FUNCNAME[0]} Failed after ${maxRetries} attempts!"
        exit 1
    fi
}

export CRIO_VERSION=1.30
cat << EOF >/etc/yum.repos.d/devel_kubic_libcontainers_stable_cri-o_${CRIO_VERSION}.repo
[isv_kubernetes_addons_cri-o_stable_v${CRIO_VERSION}]
name=CRI-O v${CRIO_VERSION} (Stable) (rpm)
type=rpm-md
baseurl=https://storage.googleapis.com/kubevirtci-crio-mirror/isv_kubernetes_addons_cri-o_stable_v${CRIO_VERSION}
gpgcheck=0
enabled=1
EOF

dnf install -y cri-o

systemctl enable --now crio

cat << EOF > /etc/containers/registries.conf
[registries.search]
registries = ['registry.access.redhat.com', 'registry.fedoraproject.org', 'quay.io', 'docker.io']

[registries.insecure]
registries = ['registry:5000']

[registries.block]
registries = []
EOF

packages_version=$(getKubernetesClosestStableVersion)
major_version=$(echo $packages_version | cut -d "." -f 2)
# Add Kubernetes release repository.
# use repodata from GCS bucket, since the release repo might not have it right after the release
# we deduce the https path from the gcs path gs://kubernetes-release/release/${version}/rpm/x86_64/
# see https://github.com/kubernetes/kubeadm/blob/main/docs/testing-pre-releases.md#availability-of-pre-compiled-release-artifacts
cat <<EOF >/etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes Release
baseurl=https://pkgs.k8s.io/core:/stable:/v1.${major_version}/rpm
enabled=1
gpgcheck=0
repo_gpgcheck=0
EOF

# Install Kubernetes CNI.
dnf install --skip-broken --nobest --nogpgcheck --disableexcludes=kubernetes -y \
    kubectl-${packages_version} \
    kubeadm-${packages_version} \
    kubelet-${packages_version} \
    kubernetes-cni

# In case the version is unstable the package manager recognizes only the closest stable version
# But it's unsafe using older kubeadm version than kubernetes version according to:
# https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/#kubeadm-s-skew-against-the-kubernetes-version
# The reason we install kubeadm using dnf is for the dependencies packages.
if [[ $version != $packages_version ]]; then
   replaceKubeBinaries
fi

kubeadm config images pull --kubernetes-version ${version}

kubectl kustomize /tmp/prometheus/grafana > /tmp/grafana-deployment.yaml.tmp
mv -f /tmp/grafana-deployment.yaml.tmp /tmp/prometheus/grafana/grafana-deployment.yaml

if [[ ${slim} == false ]]; then
    # Pre pull all images from the manifests
    for image in $(/tmp/fetch-images.sh /tmp); do
        pull_container_retry "${image}"
    done

    # Pre pull additional images from list
    for image in $(cat "/tmp/extra-pre-pull-images"); do
        pull_container_retry "${image}"
    done
fi

mkdir -p /provision

cni_manifest="/provision/cni.yaml"
cni_diff="/tmp/cni.diff"
cni_manifest_ipv6="/provision/cni_ipv6.yaml"
cni_ipv6_diff="/tmp/cni_ipv6.diff"

cp /tmp/cni.do-not-change.yaml $cni_manifest
mv /tmp/cni.do-not-change.yaml $cni_manifest_ipv6
patch $cni_manifest $cni_diff
patch $cni_manifest_ipv6 $cni_ipv6_diff

cp /tmp/local-volume.yaml /provision/local-volume.yaml

# Create drop-in config files for kubelet
# https://kubernetes.io/docs/tasks/administer-cluster/kubelet-config-file/#kubelet-conf-d
kubelet_conf_d="/etc/kubernetes/kubelet.conf.d"
mkdir -m 644 $kubelet_conf_d

# Set our custom initializations to kubelet
kubevirt_kubelet_conf="$kubelet_conf_d/50-kubevirt.conf"
cat <<EOF >$kubevirt_kubelet_conf
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
cgroupDriver: systemd
failSwapOn: false
kubeletCgroups: /systemd/system.slice
EOF

# Set only command line options not supported by config
cat <<EOT >/etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=--runtime-cgroups=/systemd/system.slice --config-dir=$kubelet_conf_d
EOT

# Enable userfaultfd for centos9 to support post-copy live migration.
# For more info: https://github.com/openshift/machine-config-operator/pull/3724
echo "vm.unprivileged_userfaultfd = 1" > /etc/sysctl.d/enable-userfaultfd.conf

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

# Delete conf files created by crio / podman
# so calico will create the interfaces by its own according the right configuration.
# See https://github.com/cri-o/cri-o/issues/2411#issuecomment-540006558
# It should happen before crio start, see https://github.com/cri-o/cri-o/issues/4276
# About podman see https://github.com/kubernetes/kubernetes/issues/107687
rm -f /etc/cni/net.d/*

systemctl daemon-reload
systemctl enable crio kubelet --now

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

echo "net.netfilter.nf_conntrack_max=1000000" >> /etc/sysctl.conf
sysctl --system

systemctl restart NetworkManager

nmcli connection modify "System eth0" \
   ipv6.method auto \
   ipv6.addr-gen-mode eui64
nmcli connection up "System eth0"

kubeadmn_patches_path="/provision/kubeadm-patches"
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

# psa configuration
cat > /etc/kubernetes/psa.yaml <<EOF
apiVersion: apiserver.config.k8s.io/v1
kind: AdmissionConfiguration
plugins:
- name: PodSecurity
  configuration:
    apiVersion: pod-security.admission.config.k8s.io/v1
    kind: PodSecurityConfiguration
    defaults:
      enforce: "privileged"
      enforce-version: "latest"
      audit: "restricted"
      audit-version: "latest"
      warn: "restricted"
      warn-version: "latest"
    exemptions:
      usernames: []
      runtimeClasses: []
      # Hopefuly this will not be needed in future. Add your favorite namespace to be ignored and your operator not broken
      # You also need to modify psa.sh
      namespaces: ["kube-system", "default", "istio-operator" ,"istio-system", "nfs-csi", "monitoring", "rook-ceph", "cluster-network-addons", "sonobuoy"]
EOF

kubeadm_raw=/tmp/kubeadm.conf
kubeadm_raw_ipv6=/tmp/kubeadm_ipv6.conf
kubeadm_manifest="/etc/kubernetes/kubeadm.conf"
kubeadm_manifest_ipv6="/etc/kubernetes/kubeadm_ipv6.conf"

envsubst < $kubeadm_raw > $kubeadm_manifest
envsubst < $kubeadm_raw_ipv6 > $kubeadm_manifest_ipv6

until ip address show dev eth0 | grep global | grep inet6; do sleep 1; done

# 1.23 has deprecated --experimental-patches /provision/kubeadm-patches/, we now mention the patch directory in kubeadm.conf
kubeadm init --config $kubeadm_manifest -v5

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

# copy network addons operator manifests
# so we can use them at cluster-up
cp -rf /tmp/cnao/ /opt/

# copy whereabouts manifests
# so we can use them at cluster-up
cp -rf /tmp/whereabouts/ /opt/

# copy Multus CNI manifests so we can use them at cluster-up
cp -rf /tmp/multus /opt/

# copy cdi manifests
cp -rf /tmp/cdi*.yaml /opt/

# copy aaq manifests
cp -rf /tmp/aaq/ /opt/

# copy kwok manifests
cp -rf /tmp/kwok /opt/

# Create a properly labelled tmp directory for testing
mkdir -p /var/provision/kubevirt.io/tests
chcon -t container_file_t /var/provision/kubevirt.io/tests
echo "tmpfs /var/provision/kubevirt.io/tests tmpfs rw,context=system_u:object_r:container_file_t:s0 0 1" >> /etc/fstab

# Cleanup the existing NetworkManager profiles so the VM instances will come
# up with the default profiles. (Base VM image includes non default settings)
rm -f /etc/sysconfig/network-scripts/ifcfg-*
nmcli connection add con-name eth0 ifname eth0 type ethernet

# Remove machine-id, allowing unique ID/s for its instances
rm -f /etc/machine-id ; touch /etc/machine-id
