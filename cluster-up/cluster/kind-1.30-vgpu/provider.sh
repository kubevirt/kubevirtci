#!/usr/bin/env bash

set -e

export CLUSTER_NAME="vgpu"


function detect_cri() {
    if podman ps >/dev/null 2>&1; then echo podman; elif docker ps >/dev/null 2>&1; then echo docker; fi
}

make -C cluster-provision/gocli cli

export CRI_BIN=${CRI_BIN:-$(detect_cri)}

function up() {
    # print hardware info for easier debugging based on logs
    echo 'Available cards'
    ${CRI_BIN} run --rm --cap-add=SYS_RAWIO quay.io/phoracek/lspci@sha256:0f3cacf7098202ef284308c64e3fc0ba441871a846022bb87d65ff130c79adb1 sh -c "lspci -k | grep -EA2 'VGA|3D'"
    echo ""

    if [ "$CI" == "true" ]; then export REGISTRY_PROXY="docker-mirror-proxy.kubevirt-prow.svc"; fi
    ./cluster-provision/gocli/build/cli run-kind vgpu \
        --with-extra-mounts=true \
        --nodes=$KUBEVIRT_NUM_NODES \
        --registry-proxy=$REGISTRY_PROXY
}

source ${KUBEVIRTCI_PATH}/cluster/kind/common.sh