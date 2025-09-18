#!/bin/bash

set -ex

ARCH=$(uname -m)

KUBEVIRTCI_SHARED_DIR=/var/lib/kubevirtci
mkdir -p $KUBEVIRTCI_SHARED_DIR
export ISTIO_VERSION=1.26.4
cat << EOF > $KUBEVIRTCI_SHARED_DIR/shared_vars.sh
#!/bin/bash
set -ex
export KUBELET_CGROUP_ARGS="--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice"
export ISTIO_VERSION=${ISTIO_VERSION}
export ISTIO_BIN_DIR="/opt/istio-${ISTIO_VERSION}/bin"
EOF
source $KUBEVIRTCI_SHARED_DIR/shared_vars.sh

# Install modules of the initrd kernel
dnf install -y "kernel-modules-$(uname -r)"

# Resize root partition
dnf install -y cloud-utils-growpart
if growpart /dev/vda 1; then
    DEVICE="/dev/vda1"
    MOUNTPOINT=$(findmnt -n -o TARGET "$DEVICE")
    FSTYPE=$(lsblk -no FSTYPE "$DEVICE")
    if [[ "$FSTYPE" == ext2 || "$FSTYPE" == ext3 || "$FSTYPE" == ext4 ]]; then
        echo "Resizing ext2/3/4 filesystem on $DEVICE..."
        resize2fs "$DEVICE"
    elif [[ "$FSTYPE" == xfs ]]; then
        echo "Resizing XFS filesystem on $DEVICE..."
        xfs_growfs "$MOUNTPOINT"
    else
        echo "Unsupported filesystem type: $FSTYPE"
        exit 1
    fi
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
  curl -L  https://github.com/istio/istio/releases/download/${ISTIO_VERSION}/istio-${ISTIO_VERSION}-linux-amd64.tar.gz -O
  tar -xvf ./istio-${ISTIO_VERSION}-linux-amd64.tar.gz --strip-components=2 -C ${ISTIO_BIN_DIR} istio-${ISTIO_VERSION}/bin/istioctl
  chmod +x "$ISTIO_BIN_DIR/istioctl"
)

dnf install -y container-selinux

dnf install -y libseccomp-devel

#openvswitch for s390x is not available from the centos default repos.
if [ "$ARCH" == "s390x" ]; then
  dnf install -y https://kojipkgs.fedoraproject.org//packages/openvswitch/2.16.0/2.fc36/s390x/openvswitch-2.16.0-2.fc36.s390x.rpm
  systemctl enable openvswitch
else
  dnf install -y centos-release-nfv-openvswitch
  dnf install -y openvswitch2.16
fi 

dnf install -y NetworkManager NetworkManager-ovs NetworkManager-config-server

# envsubst pkg is not available by default in s390x Architecture, so explicitly installing it as part of gettext
dnf install -y gettext