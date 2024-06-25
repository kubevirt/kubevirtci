#!/bin/bash -e
ARCH ?= $$(uname -m | grep -q s390x && echo s390x || echo amd64)

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

centos_version="$(cat $DIR/version | tr -d '\n')"

docker build --platform="linux/${ARCH}" --build-arg centos_version=$centos_version . -t quay.io/kubevirtci/centos9
