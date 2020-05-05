#!/usr/bin/env bash
set -e

if [ "$KUBEVIRT_WITH_CNAO" == "true" ]; then
    echo "ERROR: KUBEVIRT_WITH_CNAO=true is not supported with $KUBEVIRT_PROVIDER"
    exit 1
fi
source ${KUBEVIRTCI_PATH}/cluster/k8s-provider-common.sh
