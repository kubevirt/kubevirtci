#!/usr/bin/env bash

set -e

export CLUSTER_NAME="k3d"
export HOST_PORT=5000

function up() {
    k3d_up

    version=$(_kubectl get node k3d-$CLUSTER_NAME-server-0 -o=custom-columns=VERSION:.status.nodeInfo.kubeletVersion --no-headers)
    echo "$KUBEVIRT_PROVIDER cluster '$CLUSTER_NAME' is ready ($version)"
}

source ${KUBEVIRTCI_PATH}/cluster/k3d/common.sh
