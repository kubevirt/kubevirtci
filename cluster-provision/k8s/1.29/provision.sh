#!/bin/bash

set -ex

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

KUBEVIRTCI_SHARED_DIR=/var/lib/kubevirtci
mkdir -p $KUBEVIRTCI_SHARED_DIR
export ISTIO_VERSION=1.15.0
cat << EOF > $KUBEVIRTCI_SHARED_DIR/shared_vars.sh
#!/bin/bash
set -ex
export KUBELET_CGROUP_ARGS="--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice"
export ISTIO_VERSION=${ISTIO_VERSION}
export ISTIO_BIN_DIR="/opt/istio-${ISTIO_VERSION}/bin"
EOF
source $KUBEVIRTCI_SHARED_DIR/shared_vars.sh

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

# Install modules of the initrd kernel
dnf install -y "kernel-modules-$(uname -r)"

# Resize root partition
dnf install -y cloud-utils-growpart
if growpart /dev/vda 1; then
    resize2fs /dev/vda1
fi

dnf install -y patch

systemctl stop firewalld || :
systemctl disable firewalld || :
# Make sure the firewall is never enabled again
# Enabling the firewall destroys the iptable rules
dnf -y remove firewalld

# Required for iscsi demo to work.
dnf -y install iscsi-initiator-utils

# required for some sig-network tests
dnf -y install nftables

# for rook ceph
dnf -y install lvm2
# Convince ceph our storage is fast (not a rotational disk)
echo 'ACTION=="add|change", SUBSYSTEM=="block", KERNEL=="vd[a-z]", ATTR{queue/rotational}="0"' \
	> /etc/udev/rules.d/60-force-ssd-rotational.rules

# To prevent preflight issue related to tc not found
dnf install -y iproute-tc
# Install istioctl
export PATH="$ISTIO_BIN_DIR:$PATH"
(
  set -E
  mkdir -p "$ISTIO_BIN_DIR"
  curl "https://storage.googleapis.com/kubevirtci-istioctl-mirror/istio-${ISTIO_VERSION}/bin/istioctl" -o "$ISTIO_BIN_DIR/istioctl"
  chmod +x "$ISTIO_BIN_DIR/istioctl"
)

export CRIO_VERSION=1.28
cat << EOF >/etc/yum.repos.d/devel_kubic_libcontainers_stable_cri-o_${CRIO_VERSION}.repo
[isv_kubernetes_addons_cri-o_stable_v${CRIO_VERSION}]
name=CRI-O v${CRIO_VERSION} (Stable) (rpm)
type=rpm-md
baseurl=https://storage.googleapis.com/kubevirtci-crio-mirror/isv_kubernetes_addons_cri-o_stable_v${CRIO_VERSION}
gpgcheck=0
enabled=1
EOF

dnf install -y cri-o container-selinux

systemctl enable --now crio

dnf install -y libseccomp-devel

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

dnf install -y centos-release-nfv-openvswitch
dnf install -y openvswitch2.16

mkdir -p /provision

cni_manifest="/provision/cni.yaml"
cni_diff="/tmp/cni.diff"
cni_manifest_ipv6="/provision/cni_ipv6.yaml"
cni_ipv6_diff="/tmp/cni_ipv6.diff"

cp /tmp/cni.do-not-change.yaml $cni_manifest
mv /tmp/cni.do-not-change.yaml $cni_manifest_ipv6
patch $cni_manifest $cni_diff
patch $cni_manifest_ipv6 $cni_ipv6_diff

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
