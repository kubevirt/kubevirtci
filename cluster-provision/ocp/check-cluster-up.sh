#!/bin/bash -xe

export PARENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../.. && pwd )"
export KUBEVIRT_PROVIDER=ocp-$1
export KUBEVIRTCI_DIR="$( cd ${PARENT_DIR}/../kubevirtci && pwd)"
export KUBEVIRTCI_PROVISION_CHECK=1

function cleanup {
  cd $KUBEVIRTCI_DIR
  make cluster-down
}

pushd ${KUBEVIRTCI_DIR}
trap cleanup EXIT ERR SIGINT SIGTERM SIGQUIT
make cluster-up



