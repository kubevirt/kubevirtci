#!/usr/bin/env bash
set -e
export KUBEVIRT_PROVIDER=k8s-1.17
export WITH_CNAO=true
source ${KUBEVIRTCI_PATH}/cluster/k8s-provider-common.sh
