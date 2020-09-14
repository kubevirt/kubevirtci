#!/usr/bin/env bash
set -exuo pipefail

export REGISTRY=${REGISTRY:-docker.io}
export REPOSITORY=${REPOSITORY:-kubevirt}

full_image_tag="${REGISTRY}/${REPOSITORY}/${IMAGE_NAME}:${TAG}"

docker tag "${IMAGE_NAME}:${TAG}" "${full_image_tag}"
docker push "${full_image_tag}"
