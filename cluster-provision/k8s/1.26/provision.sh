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

function replaceKubeadmBinary() {
  dnf install -y which
  rm -f `which kubeadm`

  DOWNLOAD_DIR="/usr/bin"
  mkdir -p "$DOWNLOAD_DIR"
  RELEASE="v$version"

  ARCH="amd64"
  cd $DOWNLOAD_DIR
  curl -L --remote-name-all https://dl.k8s.io/release/${RELEASE}/bin/linux/${ARCH}/kubeadm
  chmod +x kubeadm
  cd -
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
elif grep -q "CentOS Stream 8" /etc/os-release; then
  release="centos8"
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

# Install modules of the initrd kernel
dnf install -y "kernel-modules-$(uname -r)"

# Resize root partition
dnf install -y cloud-utils-growpart
if growpart /dev/vda 1; then
    if [[ "$release" == "centos8" ]]; then
      xfs_growfs -d /
    elif [[ "$release" == "centos9" ]]; then
      resize2fs /dev/vda1
    fi
fi

dnf install -y patch

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
dnf install -y iproute-tc

# The selinux-policy package shipped with the latest Centos stream "release" can be outdated.
# Example: Centos 8 20220913.0 ships with selinux-policy-3.14.3-108, which misses crucial permissions
dnf -y update selinux-policy

# Install istioctl
export PATH="$ISTIO_BIN_DIR:$PATH"
(
  set -E
  mkdir -p "$ISTIO_BIN_DIR"
  curl "https://storage.googleapis.com/kubevirtci-istioctl-mirror/istio-${ISTIO_VERSION}/bin/istioctl" -o "$ISTIO_BIN_DIR/istioctl"
  chmod +x "$ISTIO_BIN_DIR/istioctl"
)

export CRIO_VERSION=1.26
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

echo "" >> /etc/containers/policy.json

systemctl enable --now crio

# install podman for functionality missing in crictl (tag, etc)
dnf install -y podman
dnf install -y libseccomp-devel

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

packages_version=$(getKubernetesClosestStableVersion)

# Add Kubernetes release repository.
# use repodata from GCS bucket, since the release repo might not have it right after the release
# we deduce the https path from the gcs path gs://kubernetes-release/release/${version}/rpm/x86_64/
# see https://github.com/kubernetes/kubeadm/blob/main/docs/testing-pre-releases.md#availability-of-pre-compiled-release-artifacts
cat <<EOF >/etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes Release
baseurl=https://storage.googleapis.com/kubernetes-release/release/v${packages_version}/rpm/x86_64/
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
   replaceKubeadmBinary
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
