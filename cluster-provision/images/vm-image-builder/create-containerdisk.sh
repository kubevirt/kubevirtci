#!/usr/bin/env bash
set -exuo pipefail

# dnf -y install bridge-utils libvirt virt-install qemu-kvm cloud-utils guestfs-tools

if [ "$#" -ne 1 ]; then
    echo 'Usage: create-containerdisk.sh image-directory'
    echo 'Run `create-continerdisk.sh example` to build the `example` image in the `example` folder'
    exit 1
fi

SCRIPT_PATH=$(dirname "$(realpath "$0")")

export CUSTOMIZE_IMAGE_SCRIPT=${CUSTOMIZE_IMAGE_SCRIPT:-"${SCRIPT_PATH}/customize-image.sh"}

function customize_image() {
  local source_image=$1
  local os_variant=$2
  local customized_image=$3
  local cloud_config=$4

  # Backup the VM image and pass copy of the original image
  # in case customizing script fail.
  vm_image_copy="$(dirname "${customized_image}")/copy-${source_image}"
  cp "${source_image}" "${vm_image_copy}"

  # TODO: convert this script and its dependencies to container
  ${CUSTOMIZE_IMAGE_SCRIPT} "${vm_image_copy}" "${os_variant}" "${customized_image}" "${cloud_config}"

  # Backup no longer needed.
  rm -f "${vm_image_copy}"
}

function cleanup() {
  if [ $? -ne 0 ]; then
    rm -rf "${build_directory}"
  fi

  rm -f "copy-${VM_IMAGE}"
}

export IMAGE_NAME=$1
export TAG=devel
export OS_VARIANT="$(cat ${SCRIPT_PATH}/${IMAGE_NAME}/os-variant)"
export CLOUD_CONFIG_PATH="${SCRIPT_PATH}/${IMAGE_NAME}/cloud-config"
export VM_IMAGE_URL="$(cat ${SCRIPT_PATH}/${IMAGE_NAME}/image-url)"

readonly VM_IMAGE="source-image.qcow2"
readonly build_directory="${IMAGE_NAME}_build"
readonly new_vm_image_name="provisioned-image.qcow2"

trap 'cleanup' EXIT SIGINT

pushd "${SCRIPT_PATH}"
  cleanup
  echo "Downloading the base image ..."

   if ! [ -e "${VM_IMAGE}" ]; then
    # Download base VM image
    curl -L "${VM_IMAGE_URL}" -o "${VM_IMAGE}"
  fi

  mkdir "${build_directory}"

  echo "Running the image customization ..."
  customize_image "${VM_IMAGE}" "${OS_VARIANT}" "${build_directory}/${new_vm_image_name}" "${CLOUD_CONFIG_PATH}"

  echo "Creating the containerdisk ..."
  docker build . -t ${IMAGE_NAME}:${TAG} -f - <<END
FROM scratch
ADD --chown=107:107 ${build_directory}/${new_vm_image_name} /disk/
END
popd
