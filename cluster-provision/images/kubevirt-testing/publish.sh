#!/bin/bash

function detect_cri() {
    if podman ps >/dev/null 2>&1; then
        echo podman
    elif docker ps >/dev/null 2>&1; then
        echo docker
    else
        echo "Error: no container runtime detected. Please install Podman or Docker." >&2
        exit 1
    fi
}

export CRI_BIN=${CRI_BIN:-$(detect_cri)}

${CRI_BIN} tag kubevirtci/kubevirt-testing:latest quay.io/kubevirtci/kubevirt-testing:latest
${CRI_BIN} push quay.io/kubevirtci/kubevirt-testing:latest
