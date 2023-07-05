#!/usr/bin/env bash
set -exuo pipefail

SCRIPT_PATH=$(dirname "$(realpath "$0")")
source ${SCRIPT_PATH}/common.sh
ARCH=${ARCHITECTURE:-"$(go_style_local_arch)"}
CONSOLE=${CONSOLE:-"true"}
DEBUG=${DEBUG:-"false"}

function cleanup() {
  if [ $? -ne 0 ]; then
    rm -f "${CUSTOMIZE_IMAGE_PATH}"
  fi

  virsh destroy "${DOMAIN_NAME}" || true
  undefine_vm "${DOMAIN_NAME}"
  rm -rf "${CLOUD_INIT_ISO}"
}

function undefine_vm() {
    local -r domain=$1
    virsh undefine --nvram "${domain}" || true
}

SOURCE_IMAGE_PATH=$1
OS_VARIANT=$2
CUSTOMIZE_IMAGE_PATH=$3
CLOUD_CONFIG_PATH=$4

readonly DOMAIN_NAME="provision-vm"
readonly CLOUD_INIT_ISO="cloudinit.iso"

trap 'cleanup' EXIT SIGINT

# Create cloud-init user data ISO
cloud-localds "${CLOUD_INIT_ISO}" "${CLOUD_CONFIG_PATH}"

echo "Customize image by booting a VM with
 the image and cloud-init disk
 press ctrl+] to exit"

# Check if it is native build, if true use kvm
# otherwise use emulation
if [[ ${ARCH} = $(go_style_local_arch) ]]; then
  buildconfig="--virt-type kvm"
else
  buildconfig="--arch $(linux_style_arch_name ${ARCH})"
fi

if [[ ${DEBUG} = "true" ]]; then
  buildconfig="${buildconfig} --debug --serial file,path=/tmp/provision-vm-console.log"
fi

if [[ ${CONSOLE} = "false" ]]; then
  consoleconfig="--noautoconsole --wait 120"
else
  consoleconfig=""
fi

virt-install \
  --memory 2048 \
  --vcpus 2 \
  --name $DOMAIN_NAME \
  --disk "${SOURCE_IMAGE_PATH}",device=disk \
  --disk "${CLOUD_INIT_ISO}",device=cdrom \
  --os-type Linux \
  --os-variant "${OS_VARIANT}" \
  --graphics none \
  --network default \
  --import \
  ${buildconfig} \
  ${consoleconfig}

if [[ ${DEBUG} = "true" ]]; then
  cat /tmp/provision-vm-console.log
fi

# Stop VM
virsh destroy $DOMAIN_NAME || true

# Prepare VM image
virt-sysprep -d $DOMAIN_NAME --operations machine-id,bash-history,logfiles,tmp-files,net-hostname,net-hwaddr

# Remove VM
undefine_vm "${DOMAIN_NAME}"

# Convert image
qemu-img convert -c -O qcow2 "${SOURCE_IMAGE_PATH}" "${CUSTOMIZE_IMAGE_PATH}"
