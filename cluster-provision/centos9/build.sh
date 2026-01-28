#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

centos_version="$(cat $DIR/version | tr -d '\n')"

source "${DIR}/../../hack/detect_cri.sh"
export CRI_BIN=${CRI_BIN:-$(detect_cri)}

${CRI_BIN} build --build-arg BUILDARCH=$(uname -m) --build-arg centos_version=$centos_version . -t quay.io/kubevirtci/centos9
