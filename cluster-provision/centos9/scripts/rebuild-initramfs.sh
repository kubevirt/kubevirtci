#!/bin/bash
# Script to rebuild initramfs with custom udev rules for persistent interface naming
set -ex

UDEV_RULES_FILE="/scripts/71-persistent-net-names.rules"

if [ ! -f "$UDEV_RULES_FILE" ]; then
    echo "Error: udev rules file not found at $UDEV_RULES_FILE"
    exit 1
fi

# Extract the current initramfs
mkdir -p /tmp/initramfs
cd /tmp/initramfs
zcat /initrd.img | cpio -idmv

# Create udev rules directory if it doesn't exist
mkdir -p etc/udev/rules.d

# Copy our custom udev rules
cp "$UDEV_RULES_FILE" etc/udev/rules.d/71-persistent-net-names.rules

# Rebuild initramfs
find . | cpio -o -H newc | gzip > /initrd.img.new

# Backup original and replace
mv /initrd.img /initrd.img.orig
mv /initrd.img.new /initrd.img

# Cleanup
cd /
rm -rf /tmp/initramfs

echo "Initramfs rebuilt successfully with persistent network naming rules"

