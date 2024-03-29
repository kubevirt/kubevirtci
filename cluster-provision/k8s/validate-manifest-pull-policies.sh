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
	 quay.io/kubevirtci/check-image-pull-policies@sha256:c942d3a4a17f1576f81eba0a5844c904d496890677c6943380b543bbf2d9d1be \
            --manifest-source=/manifests \
            --dry-run=false \
            --verbose=false
}

main "$@"
