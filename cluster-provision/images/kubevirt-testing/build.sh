#!/bin/bash
export DOCKER_CLI_EXPERIMENTAL=enabled

SCRIPT_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")"
    pwd
)"

source ${SCRIPT_DIR}/common.sh

# setup buildx environment
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
docker buildx create --use

# build multi-arch images
for ARCH in ${ARCHS}; do
        baseimage=$(grep "${ARCH}" ${SCRIPT_DIR}/BASEIMAGE | cut -d= -f2)
        docker buildx build --load --platform linux/${ARCH}  -t ${DOCKER_PREFIX}/${REPOSITORY}/kubevirt-testing-${ARCH} --build-arg BASEIMAGE=${baseimage} ${SCRIPT_DIR}
done
