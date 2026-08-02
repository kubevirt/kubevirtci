#!/bin/bash

set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

CENTOS_VERSION="$(cat $DIR/version)"

source "${DIR}/../../hack/detect_cri.sh"
export CRI_BIN=${CRI_BIN:-$(detect_cri)}

${CRI_BIN} build --build-arg BUILDARCH=$(uname -m) --build-arg CENTOS_VERSION=$CENTOS_VERSION . -t quay.io/kubevirtci/centos9
