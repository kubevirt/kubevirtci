#!/usr/bin/env bash
set -exuo pipefail

if [ "$#" -ne 1 ]; then
    echo "Usage: create-containerdisk.sh image-directory"
    echo "Run `create-continerdisk.sh example` to build the `example` image in the `example` folder"
fi

SCRIPT_PATH=$(dirname "$(realpath "$0")")
source ${SCRIPT_PATH}/common.sh
ARCH="${ARCHITECTURE:-"$(go_style_local_arch)"}"

export CUSTOMIZE_IMAGE_SCRIPT=${CUSTOMIZE_IMAGE_SCRIPT:-"${SCRIPT_PATH}/customize-image.sh"}

function download_base_image() {
    local -r arch=$1
    local -r image_name=$2

    local url
    local file_name
    if [[ ${ARCH} = "amd64" ]]; then
        url="$(cat ${SCRIPT_PATH}/${IMAGE_NAME}/image-url)"
        file_name="${image_name}-image.qcow2"
    else
        image_url_path=${SCRIPT_PATH}/${IMAGE_NAME}/image-url-${ARCH}
        url="$(cat ${image_url_path})"
        file_name="${image_name}-image-${ARCH}.qcow2"
    fi

    if ! [ -e "${file_name}" ]; then
        # Download base VM image
        wget -q "${url}" -O "${file_name}"
    fi

    echo "${file_name}"
}

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

function build_container() {
    local -r build_directory=$1
    local -r new_vm_image_name=$2
    local -r arch=$3
    local -r image_name=$4

    if [[ $arch = $(go_style_local_arch) ]]; then
        podman_build="podman build . -t ${image_name}:${arch}"
    else
        podman_build="podman build --platform "linux/${arch}" . -t ${image_name}:${arch}"
    fi
    $podman_build -f - <<END
FROM scratch
ADD --chown=107:107 ${build_directory}/${new_vm_image_name} /disk/
END
}

function cleanup() {
    if [ $? -ne 0 ]; then
        rm -rf "${build_directory}"
    fi

    rm -f "copy-${customized_image}"
}

trap 'cleanup' EXIT SIGINT

export IMAGE_NAME=$1

if [ -f "${SCRIPT_PATH}/${IMAGE_NAME}/create-image.sh" ]; then
    export TAG="devel"
    readonly customized_image="customized-image.qcow2"
    readonly build_directory="${SCRIPT_PATH}/${IMAGE_NAME}/build"

    trap 'cleanup' EXIT SIGINT
    mkdir -p "${build_directory}"

    pushd "${SCRIPT_PATH}/${IMAGE_NAME}"
      cleanup
      echo "Creating the image"
      ./create-image.sh "${build_directory}/${customized_image}"

      echo "Creating the containerdisk ..."
      podman build . -t ${IMAGE_NAME}:${TAG} -f - <<END
FROM scratch
ADD --chown=107:107 build/${customized_image} /disk/
END
    popd
else
    export OS_VARIANT="$(cat ${SCRIPT_PATH}/${IMAGE_NAME}/os-variant)"
    export CLOUD_CONFIG_PATH="${SCRIPT_PATH}/${IMAGE_NAME}/cloud-config"
    build_directory="${IMAGE_NAME}_build"
    customized_image="customized-image.qcow2"

    trap 'cleanup' EXIT SIGINT

    pushd "${SCRIPT_PATH}"
      cleanup
      echo "Downloading the base image ..."
      base_image=$(download_base_image "${ARCH}" "${IMAGE_NAME}")

      mkdir -p "${build_directory}"

      echo "Running the image customization ..."
      ARCHITECTURE="${ARCH}" customize_image "${base_image}" "${OS_VARIANT}" "${build_directory}/${customized_image}" "${CLOUD_CONFIG_PATH}"

      echo "Creating the containerdisk ..."
      build_container "${build_directory}" "${customized_image}" "${ARCH}" "${IMAGE_NAME}"
    popd
fi
