#!/bin/bash

set -ex

setenforce 0

# TODO: Can this be moved to privision script?
# Having these packages installed does not have impact on the regular deployments
yum install -y dpdk dpdk-devel

# Container image is creted using the build scripts in repo https://github.com/krsacme/container-builds-scripts
# By using the prebuilt driver and rpm, the time is reduced in CI to build it
KUBEVIRT_OVSDPDK_BUILD=${KUBEVIRT_OVSDPDK_BUILD:-"false"}
KUBEVIRT_OVSDPDK_HELPER_IMAGE=${KUBEVIRT_OVSDPDK_HELPER_IMAGE:-"quay.io/krsacme/kubevirt-ovsdpdk-helpers:latest"}
if [[ "$KUBEVIRT_OVSDPDK_BUILD" == "false" ]]; then
    echo "Copying driver and rpm from helper image..."
    FILES="/tmp/files"
    mkdir -p $FILES
    docker run -v $FILES:/host $KUBEVIRT_OVSDPDK_HELPER_IMAGE

    modprobe uio
    insmod $FILES/igb_uio.ko
    yum install -y  $FILES/openvswitch-2.13*

    # /proc/self/pagemap permission issue, running as root for now
    sed -i 's/^OVS_USER_ID/#OVS_USER_ID/g' /etc/sysconfig/openvswitch
    systemctl enable --now openvswitch
    exit 0
fi

###############################################################################
echo "Building driver and rpm source locally..."

yum config-manager --set-enabled PowerTools
yum install -y wget make gcc numactl-devel kernel-devel git autoconf automake libtool  libcap-ng-devel python3 rpm-build openssl-devel unbound unbound-devel selinux-policy-devel graphviz gcc-c++ desktop-file-utils procps-ng python3-devel libpcap-devel libmnl-devel  glibc groff python3-sphinx libibverbs libibverbs-devel elfutils-libelf-devel

cd $HOME
git clone --depth 1 --single-branch --branch branch-2.13 https://github.com/openvswitch/ovs.git
cd ovs/
./boot.sh
./configure --with-dpdk=/usr/share/dpdk/x86_64-default-linux-gcc/ --prefix=/usr --localstatedir=/var --sysconfdir=/etc

# virtio driver does not support MQ
sed -i 's/ETH_MQ_RX_RSS/ETH_MQ_RX_NONE/g' lib/netdev-dpdk.c

make rpm-fedora RPMBUILD_OPT="--with dpdk --without check"
rpm -iv rpm/rpmbuild/RPMS/x86_64/openvswitch-2.13*

# /proc/self/pagemap permission issue, running as root for now
sed -i 's/^OVS_USER_ID/#OVS_USER_ID/g' /etc/sysconfig/openvswitch
systemctl enable --now openvswitch


cd $HOME
wget https://fast.dpdk.org/rel/dpdk-19.11.tar.xz
tar xf dpdk-19.11.tar.xz
export DPDK_DIR=$PWD/dpdk-19.11
cd $DPDK_DIR
export DPDK_TARGET=x86_64-native-linuxapp-gcc
export DPDK_BUILD=$DPDK_DIR/build
make config T=$DPDK_TARGET

# igb_uio.ko kernel module is required from DPDK build
# Reduce the build time, by reducing unwanted libraries
sed -i -E 's/(CONFIG_RTE.*=)y/\1n/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_EAL_IGB_UIO=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_LIBRTE_EAL=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_EAL_VFIO=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_ARCH_X86_64=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_ARCH_X86=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_ARCH_64=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_TOOLCHAIN_GCC=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_EXEC_ENV_LINUX=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_EXEC_ENV_LINUXAPP=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_EAL_NUMA_AWARE_HUGEPAGES=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_LIBRTE_PCI=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_LIBRTE_KVARGS=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_LIBRTE_ETHER=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_ETHDEV_RXTX_CALLBACKS=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_BACKTRACE=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_LIBRTE_NET=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_LIBRTE_MBUF=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_LIBRTE_MEMPOOL=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_LIBRTE_RING=).*/\1y/g' $DPDK_BUILD/.config
sed -i -E 's/(CONFIG_RTE_LIBRTE_METER=).*/\1y/g' $DPDK_BUILD/.config

cd $DPDK_BUILD
make T=$DPDK_TARGET

modprobe uio
insmod $DPDK_BUILD/kmod/igb_uio.ko
