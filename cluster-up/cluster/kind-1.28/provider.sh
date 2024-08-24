#!/usr/bin/env bash

set -e

export CLUSTER_NAME="k8s-1.28"

make -C cluster-provision/gocli cli

function up() {
    if [ "$CI" == "true" ]; then export REGISTRY_PROXY="docker-mirror-proxy.kubevirt-prow.svc"; fi
    ./cluster-provision/gocli/build/cli run-kind k8s-1.28 \
        --with-extra-mounts=true \
        --nodes=$KUBEVIRT_NUM_NODES \
        --registry-proxy=$REGISTRY_PROXY
}

source ${KUBEVIRTCI_PATH}/cluster/kind/common.sh