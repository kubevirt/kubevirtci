#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

source "${SCRIPT_DIR}/../../../hack/detect_cri.sh"
export CRI_BIN=${CRI_BIN:-$(detect_cri)}

${CRI_BIN} build -t kubevirtci/kubevirt-testing:latest -f ${SCRIPT_DIR}/Dockerfile ${SCRIPT_DIR}
