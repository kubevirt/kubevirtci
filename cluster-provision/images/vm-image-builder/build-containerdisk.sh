#!/usr/bin/env bash
set -exuo pipefail

SCRIPT_PATH=$(dirname "$(realpath "$0")")

function cleanup() {
  if [ $? -ne 0 ]; then
    rm -f "${tar_file}"
    docker image rm -f "${full_image_tag}"
  fi

  rm -rf "${temp_dir}"
}

IMAGE_NAME=$1
TAG=$2
VM_IMAGE_FILE=$3

template_dockerfile="${SCRIPT_PATH}/Dockerfile.containerdisk"

parent_dir=$(dirname "$VM_IMAGE_FILE")
tar_file="${parent_dir}/${IMAGE_NAME}-${TAG}.tar"
full_image_tag="${IMAGE_NAME}:${TAG}"

trap 'cleanup' EXIT SIGINT

temp_dir=$(mktemp -d -p /tmp -t build.container-disk.XXXX)
cp "${template_dockerfile}" "${temp_dir}/Dockerfile"
cp "${VM_IMAGE_FILE}" "${temp_dir}/image"

pushd "${temp_dir}"
  docker build -t "${full_image_tag}" .
popd

docker save --output "${tar_file}" "${full_image_tag}"

echo "Container image saved as tar file at: ${tar_file}"
