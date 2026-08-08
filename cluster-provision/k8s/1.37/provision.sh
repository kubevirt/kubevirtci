#!/bin/bash

# This file is part of the KubeVirt project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Copyright the KubeVirt Authors.

set -eux -o pipefail

ARCH=$(uname -m)

KUBEVIRTCI_SHARED_DIR=/var/lib/kubevirtci
mkdir -p $KUBEVIRTCI_SHARED_DIR
export ISTIO_VERSION=1.30.2
cat << EOF > $KUBEVIRTCI_SHARED_DIR/shared_vars.sh
#!/hint/bash
export KUBELET_CGROUP_ARGS="--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice"
export ISTIO_VERSION=${ISTIO_VERSION}
export ISTIO_BIN_DIR="/opt/istio-${ISTIO_VERSION}/bin"
EOF
source $KUBEVIRTCI_SHARED_DIR/shared_vars.sh

curl()( { set +x; } 2>/dev/null; command curl --fail-with-body --retry 2 -L "$@" )

# Install modules of the initrd kernel.
KERNEL_RELEASE="$(uname -r)"
dnf install -y "kernel-modules-${KERNEL_RELEASE}" "kernel-devel-${KERNEL_RELEASE}" "kernel-modules-extra-${KERNEL_RELEASE}"

# Resize root partition
dnf install -y cloud-utils-growpart

ROOT_PART_NUM='2'
if growpart /dev/vda "${ROOT_PART_NUM}"; then
    DEVICE="/dev/vda${ROOT_PART_NUM}"
    MOUNTPOINT=$(findmnt -n -o TARGET "$DEVICE")
    FSTYPE=$(lsblk -no FSTYPE "$DEVICE")
    case "${FSTYPE}" in
        ext[234])
            echo "Resizing ext2/3/4 filesystem on $DEVICE..."
            resize2fs "$DEVICE"
        ;;
        xfs)
            echo "Resizing XFS filesystem on $DEVICE..."
            xfs_growfs "$MOUNTPOINT"
        ;;
        *)
            echo "[ERROR] Unsupported filesystem type: $FSTYPE" >&2
            exit 1
        ;;
    esac
fi

dnf install -y patch pciutils

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
cat > /etc/udev/rules.d/60-force-ssd-rotational.rules <<EOF
ACTION=="add|change", SUBSYSTEM=="block", KERNEL=="vd[a-z]", ATTR{queue/rotational}="0"
EOF

# To prevent preflight issue related to tc not found
dnf install -y iproute-tc
# Install istioctl
export PATH="$ISTIO_BIN_DIR:$PATH"
(
  set -E
  mkdir -p "$ISTIO_BIN_DIR"
  curl -O  https://github.com/istio/istio/releases/download/${ISTIO_VERSION}/istio-${ISTIO_VERSION}-linux-amd64.tar.gz
  tar -xvf ./istio-${ISTIO_VERSION}-linux-amd64.tar.gz --strip-components=2 -C ${ISTIO_BIN_DIR} istio-${ISTIO_VERSION}/bin/istioctl
  chmod +x "$ISTIO_BIN_DIR/istioctl"
)

dnf install -y container-selinux

dnf install -y libseccomp-devel

dnf install -y centos-release-nfv-openvswitch
dnf install -y openvswitch3.5

dnf install -y NetworkManager NetworkManager-ovs NetworkManager-config-server

# NetworkManager-config-server sets no-auto-default=* which prevents auto-DHCP
# on unconfigured interfaces. CentOS 9 has ifcfg-eth0 from cloud-init but
# CentOS 10 uses keyfile format and has no persistent connection profile.
cat > /etc/NetworkManager/system-connections/eth0.nmconnection << ETHEOF
[connection]
id=eth0
type=ethernet
interface-name=eth0
autoconnect=true

[ipv4]
method=auto

[ipv6]
method=auto
ETHEOF
chmod 600 /etc/NetworkManager/system-connections/eth0.nmconnection

# envsubst pkg is not available by default in s390x Architecture, so explicitly installing it as part of gettext
dnf install -y gettext
