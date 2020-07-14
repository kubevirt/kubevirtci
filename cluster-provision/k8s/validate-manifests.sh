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
    (
        cd "$DIR/../tools"
        docker build -f check-image-pull-policies/Dockerfile -t kubevirtci/check-image-pull-policies .
    )
    docker run -it --rm -v "$manifest_dir:/manifests:Z" kubevirtci/check-image-pull-policies

}

main "$@"