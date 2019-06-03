#!/bin/bash

SCRIPT_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")"
    pwd
)"

docker build -t kubevirtci/kubevirt-testing:latest -f ${SCRIPT_DIR}/Dockerfile ${SCRIPT_DIR}
