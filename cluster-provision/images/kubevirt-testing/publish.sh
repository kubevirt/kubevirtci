#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

source "${SCRIPT_DIR}/../../../hack/detect_cri.sh"
export CRI_BIN=${CRI_BIN:-$(detect_cri)}

${CRI_BIN} tag kubevirtci/kubevirt-testing:latest quay.io/kubevirtci/kubevirt-testing:latest
${CRI_BIN} push quay.io/kubevirtci/kubevirt-testing:latest
