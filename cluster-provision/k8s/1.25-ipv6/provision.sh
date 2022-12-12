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

if grep -q "CentOS Stream 9" /etc/os-release; then
  release="centos9"
elif grep -q "CentOS Stream 8" /etc/os-release; then
  release="centos8"
else
  echo "ERROR: Could not recognize guest OS"
  exit 1
fi

cni_diff="/tmp/cni.diff"
ipv6_dualstack=true
if [[ ${networkstack} == ipv6 ]]; then
    cni_diff="/tmp/cni_ipv6.diff"
    ipv6_dualstack=false
fi

KUBEVIRTCI_SHARED_DIR=/var/lib/kubevirtci
mkdir -p $KUBEVIRTCI_SHARED_DIR
cat << EOF > $KUBEVIRTCI_SHARED_DIR/shared_vars.sh
#!/bin/bash
set -ex
export KUBELET_CGROUP_ARGS="--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice"
export ISTIO_VERSION=1.15.0
export ISTIO_BIN_DIR=/opt/istio-$ISTIO_VERSION/bin
export KUBEVIRTCI_DUALSTACK=$ipv6_dualstack
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
# Install istioctl
export PATH=$ISTIO_BIN_DIR:$PATH
(
  set -E
  mkdir -p $ISTIO_BIN_DIR
  curl https://storage.googleapis.com/kubevirtci-istioctl-mirror/istio-$ISTIO_VERSION/bin/istioctl -o $ISTIO_BIN_DIR/istioctl
  chmod +x $ISTIO_BIN_DIR/istioctl
)

export CRIO_VERSION=1.25
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
if [[ "$release" == "centos8" ]]; then
    dnf install -y cri-o containers-common-1-23.module_el8.7.0+1106+45480ee0.x86_64
elif [[ "$release" == "centos9" ]]; then
    dnf install -y cri-o
fi

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

kubeadm config images pull --kubernetes-version ${version}

dnf install -y centos-release-nfv-openvswitch
dnf install -y openvswitch2.16

mkdir -p /provision
cni_manifest="/provision/cni.yaml"
mv /tmp/cni.do-not-change.yaml $cni_manifest
patch $cni_manifest $cni_diff

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
