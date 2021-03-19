#!/usr/bin/env bash

if [ -z "$KUBEVIRTCI_PATH" ]; then
    KUBEVIRTCI_PATH="$(
        cd "$(dirname "$BASH_SOURCE[0]")/"
        echo "$(pwd)/"
    )"
fi

if [ -z "$1" ]; then
    echo $0 snashot-name
    echo save a qemu-based provider to a named snapshot
    exit 1
fi

source ${KUBEVIRTCI_PATH}hack/common.sh
source ${KUBEVIRTCI_CLUSTER_PATH}/$KUBEVIRT_PROVIDER/provider.sh
save "$1"
