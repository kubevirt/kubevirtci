#!/bin/bash
export DOCKER_CLI_EXPERIMENTAL=enabled

SCRIPT_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")"
    pwd
)"

source ${SCRIPT_DIR}/common.sh

# set tags and push multiarch images
for ARCH in ${ARCHS}; do
        docker push ${DOCKER_PREFIX}/${REPOSITORY}/kubevirt-testing-${ARCH}
        manifest+=("${DOCKER_PREFIX}/${REPOSITORY}/kubevirt-testing-${ARCH}:latest")
done

# create and push manifest
docker manifest create --amend ${DOCKER_PREFIX}/${REPOSITORY}/kubevirt-testing:latest "${manifest[@]}"
set -x; for ARCH in ${ARCHS}; do docker manifest annotate --arch ${ARCH} ${DOCKER_PREFIX}/${REPOSITORY}/kubevirt-testing:latest ${DOCKER_PREFIX}/${REPOSITORY}/kubevirt-testing-${ARCH}:latest; done
docker manifest push ${DOCKER_PREFIX}/${REPOSITORY}/kubevirt-testing:latest
