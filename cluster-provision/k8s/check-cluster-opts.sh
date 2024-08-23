#!/bin/bash

set -exuo pipefail

make -C ../gocli container

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

provision_dir="$1"
provider="${provision_dir}"

function cleanup() {
    cd "$DIR" && cd ../..
    make cluster-down
}

export KUBEVIRTCI_GOCLI_CONTAINER=quay.io/kubevirtci/gocli:latest
# check cluster-up
(
    ksh="./cluster-up/kubectl.sh"
    cd "$DIR" && cd ../..
    export KUBEVIRTCI_PROVISION_CHECK=1
    export KUBEVIRT_PROVIDER="k8s-${provider}"
    export KUBEVIRT_MEMORY_SIZE=5520M

    # Test KSM and Swap
    export KUBEVIRT_KSM_ON="true"
    export KUBEVIRT_KSM_SLEEP_BETWEEN_SCANS_MS=20
    export KUBEVIRT_KSM_PAGES_TO_SCAN=10

    export KUBEVIRT_SWAP_ON="true"
    export KUBEVIRT_SWAP_SIZE_IN_GB=1

    trap cleanup EXIT ERR SIGINT SIGTERM SIGQUIT
    make cluster-up
    ${ksh} get nodes
    make cluster-down

    export KUBEVIRT_KSM_ON="false"
    export KUBEVIRT_SWAP_ON="false"

    # Test ETCD in memory
    export KUBEVIRT_WITH_ETC_IN_MEMORY="true"
    export KUBEVIRT_WITH_ETC_CAPACITY="1024M"
    make cluster-up
    ${ksh} get nodes
    make cluster-down

    # Test NFS CSI
    export KUBEVIRT_DEPLOY_NFS_CSI="true"
    make cluster-up
    ${ksh} get nodes
    make cluster-down

    export KUBEVIRT_DEPLOY_NFS_CSI="false"

    # Test rook ceph
    export KUBEVIRT_STORAGE="rook-ceph-default"
    make cluster-up
    ${ksh} get nodes
    make cluster-down
)