#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

centos_version="$(cat $DIR/version | tr -d '\n')"

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

${CRI_BIN} build --build-arg BUILDARCH=$(uname -m) --build-arg centos_version=$centos_version . -t quay.io/kubevirtci/centos9
