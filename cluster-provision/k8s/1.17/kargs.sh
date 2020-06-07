#!/bin/bash

set -ex

# OvS-DPDK requires latest kernel
yum update -y --nobest kernel

# Increase the Hugepages count so that DPDK can be enabled in OvS which requires 1GB of hugepages
# Add iommu support to enable PMD drivers
GRUB_FILE=/etc/default/grub
MODIFIED=0
if ! grep -q iommu $GRUB_FILE; then
    sed -i 's/${GRUB_CMDLINE_LINUX}/${GRUB_CMDLINE_LINUX} iommu=pt intel_iommu=on/g' $GRUB_FILE
    MODIFIED=1
fi

if grep -q 'hugepagesz=2M' $GRUB_FILE && ! grep -q 'hugepages=2048' $GRUB_FILE; then
    sed -i 's/ hugepages=[0-9]*/ hugepages=2048/g' $GRUB_FILE
    MODIFIED=1
fi

if [[ $MODIFIED == "1" ]]; then
    grub2-mkconfig -o /boot/grub2/grub.cfg
fi
