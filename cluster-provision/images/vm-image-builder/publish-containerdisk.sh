#!/usr/bin/env bash
set -exuo pipefail

function detect_cri() {
    if podman ps >/dev/null 2>&1; then echo "podman"; elif docker ps >/dev/null 2>&1; then echo "docker"; fi
}

export CRI_BIN=${CRI_BIN:-$(detect_cri)}

if [ "$#" -ne 2 ]; then
    echo "Usage: publish-containerdisk.sh <image-directory> <image-target-name> "
    echo "Run `publish-continerdisk.sh example quay.io/kubevirtci/example:mytag` to push the local `example:devel` image to `quay.io/kubevirtci/example:mytag`."
fi

export IMAGE_NAME=$1
export TAG=devel
export FULL_IMAGE_NAME=$2

${CRI_BIN} tag "${IMAGE_NAME}:${TAG}" "${FULL_IMAGE_NAME}"
${CRI_BIN} push "${FULL_IMAGE_NAME}"
