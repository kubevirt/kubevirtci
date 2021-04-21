#!/usr/bin/env bash
set -exuo pipefail

if [ "$#" -ne 2 ]; then
    echo "Usage: publish-containerdisk.sh <image-directory> <image-target-name> "
    echo "Run `publish-continerdisk.sh example quay.io/kubevirtci/example:mytag` to push the local `example:devel` image to `quay.io/kubevirtci/example:mytag`."
fi

export IMAGE_NAME=$1
export TAG=devel
export FULL_IMAGE_NAME=$2

docker tag "${IMAGE_NAME}:${TAG}" "${FULL_IMAGE_NAME}"
docker push "${FULL_IMAGE_NAME}"
