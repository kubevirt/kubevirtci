#!/bin/bash

set -exuo pipefail

function usage {
    cat <<EOF
"${BASH_SOURCE[0]} [-c|--no-cleanup]" - spin up kubevirtci cluster and check for readiness

    env vars required:
        \${version}: version of kubevirtci cluster to spin up

    --no-cleanup: do not spin down cluster after testing

EOF
}

cleanup=1
while [ "$#" -gt 0 ]; do
    case "$1" in
        -c|--no-cleanup)
            cleanup=0
            shift
            ;;
        *)
            usage
            return 1
            ;;
    esac
done

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function cleanup {
    cd "$DIR" && cd ../..
    make cluster-down
}

# check cluster-up
(
    ksh="./cluster-up/kubectl.sh"
    cd "$DIR" && cd ../..
    export KUBEVIRTCI_PROVISION_CHECK=1
    export KUBEVIRT_PROVIDER="k8s-${version}"
    export KUBEVIRT_NUM_NODES=2
    if [ "$cleanup" -ne 0 ]; then
        trap cleanup EXIT ERR SIGINT SIGTERM SIGQUIT
    fi
    bash -x ./cluster-up/up.sh
    ${ksh} wait --for=condition=Ready pod --all
    ${ksh} wait --for=condition=Ready pod -n kube-system --all
    ${ksh} get nodes
    ${ksh} get pods -A
)
