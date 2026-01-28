#!/bin/bash

set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ksh="$(cd "$DIR/../.." && pwd)/cluster-up/kubectl.sh"

source "${DIR}/../../hack/detect_cri.sh"
export CRI_BIN=${CRI_BIN:-$(detect_cri)}

function usage() {
    cat <<EOF
Usage: $0

    Dumps the pod yamls of all pods from the just provisioned k8s cluster into a folder and checks the ImagePullPolicy
    values against the conditions for _not always pulling images_ as described here:
    https://kubernetes.io/docs/concepts/containers/images/#updating-images

    (Not yet enabled) Exits with non-zero exit code if the check fails

EOF
}

function main() {

    manifest_dir=$(mktemp -d)
    trap 'rm -rf $manifest_dir' SIGINT SIGTERM EXIT
    for namespace in $(${ksh} get namespaces --no-headers --output=custom-columns=:.metadata.name); do
        if [ "$namespace" == 'kube-system' ]; then
            continue
        fi
        mkdir -p "$manifest_dir/$namespace"
        for pod in $(${ksh} get pods --namespace "$namespace" --no-headers --output=custom-columns=:.metadata.name); do
            (${ksh} get pod --output=yaml --namespace "$namespace" "$pod") >"$manifest_dir/$namespace/$pod.yaml"
        done
    done

    echo "Checking $manifest_dir"
    # TODO: for now we disable (via --dry-run) the non zero exit code in case of failure here to give the teams some time to fix the policies
    ${CRI_BIN} run --rm -v "$manifest_dir:/manifests:Z" \
      quay.io/kubevirtci/check-image-pull-policies@sha256:c942d3a4a17f1576f81eba0a5844c904d496890677c6943380b543bbf2d9d1be \
        --manifest-source=/manifests \
        --dry-run=true \
        --verbose=false
}

main "$@"
