#!/usr/bin/env bash
set -exuo pipefail

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

export IMAGE_NAME=${IMAGE_NAME:-example-fedora}
export TAG=${TAG:-32}
export OS_VARIANT=${OS_VARIANT:-fedora31}
export CLOUD_CONFIG_PATH=${CLOUD_CONFIG_PATH:-"${SCRIPT_PATH}/example/cloud-config"}
export VM_IMAGE_URL=${VM_IMAGE_URL:-$(cat "${SCRIPT_PATH}/example/image-url")}

readonly VM_IMAGE="source-image.qcow2"
readonly build_directory="${IMAGE_NAME}_build"
readonly new_vm_image_name="provisioned-image.qcow2"

trap 'cleanup' EXIT SIGINT

pushd "${SCRIPT_PATH}"
  cleanup

   if ! [ -e "${VM_IMAGE}" ]; then
    # Download base VM image
    curl -L "${VM_IMAGE_URL}" -o "${VM_IMAGE}"
  fi

  mkdir "${build_directory}"

  customize_image "${VM_IMAGE}" "${OS_VARIANT}" "${build_directory}/${new_vm_image_name}" "${CLOUD_CONFIG_PATH}"

  ${SCRIPT_PATH}/build-containerdisk.sh "${IMAGE_NAME}" "${TAG}" "${build_directory}/${new_vm_image_name}"

popd
