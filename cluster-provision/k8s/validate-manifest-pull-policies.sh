#!/bin/bash

set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function usage {
    cat <<EOF
Usage: $0 <k8s-cluster-dir>

    Checks the input that is expected to be a deployment directory containing a directory called 'manifests'.

    Exits with non-zero exit code if the check fails

EOF
}

function main {

    if [ "$#" -lt 1 ]; then
        usage
        exit 1
    fi

    manifest_dir="$DIR/$1/manifests"
    echo "Checking $manifest_dir"
    docker run --rm -v "$manifest_dir:/manifests:Z" \
        kubevirtci/check-image-pull-policies@sha256:118c4828afa52e58fc07663f400a357764cc1e7432ab56c439bb5c0b4b11b4dc \
            --manifest-source=/manifests \
            --dry-run=false \
            --verbose=false
}

main "$@"