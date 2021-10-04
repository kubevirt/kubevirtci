#!/usr/bin/env bash
set -exuo pipefail

function cleanup() {
  if [ $? -ne 0 ]; then
    rm -f "${CUSTOMIZE_IMAGE_PATH}"
  fi

  rm -rf "${CLOUD_INIT_ISO}"
  virsh destroy "${DOMAIN_NAME}" || true
  virsh undefine "${DOMAIN_NAME}" || true
}

VIRT_INSTALL_WAIT_INTERVAL=${VIRT_INSTALL_WAIT_INTERVAL:-5}
VIRT_INSTALL_TIMEOUT=${VIRT_INSTALL_TIMEOUT:-"$((100*$VIRT_INSTALL_WAIT_INTERVAL))"}

function wait_for_install_to_complete() {
  local -r timeout_seconds=$1
  local -r interval_seconds=$2
  local count=0
  local -r retries=$((timeout_seconds/interval_seconds))
  until [[ $(virsh list --state-running --name | grep $DOMAIN_NAME) == "" || $count -gt $retries ]]; do
    sleep $interval_seconds
    count=$((count + 1))
  done
  if [[ $count -gt $retries ]]; then
    echo "VM '$DOMAIN_NAME' is still in "$(virsh list --all |grep $DOMAIN_NAME |awk '{print $3}')" state after waiting for "$(expr $count \* $retries) "seconds"
    return 1
  else
    echo "done"
  fi
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
virt-install \
  --memory 2048 \
  --vcpus 2 \
  --name $DOMAIN_NAME \
  --disk "${SOURCE_IMAGE_PATH}",device=disk \
  --disk "${CLOUD_INIT_ISO}",device=cdrom \
  --os-type Linux \
  --os-variant "${OS_VARIANT}" \
  --virt-type kvm \
  --graphics none \
  --network default \
  --import

wait_for_install_to_complete $VIRT_INSTALL_TIMEOUT $VIRT_INSTALL_WAIT_INTERVAL

# Stop VM
virsh destroy $DOMAIN_NAME || true

# Prepare VM image
virt-sysprep -d $DOMAIN_NAME --operations machine-id,bash-history,logfiles,tmp-files,net-hostname,net-hwaddr

# Remove VM
virsh undefine $DOMAIN_NAME

# Convert image
qemu-img convert -c -O qcow2 "${SOURCE_IMAGE_PATH}" "${CUSTOMIZE_IMAGE_PATH}"
