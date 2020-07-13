#!/bin/bash

set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function usage {
    cat <<EOF
Usage: $0 <k8s-cluster-dir>

    Checks the input that is expected to be a list of container image indentifiers each being a fixed version
    image.

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
        docker build -f check_image_pull_policy/Dockerfile -t kubevirtci-tools/check_image_pull_policy .
    )
    docker run -it --rm -v "$manifest_dir:/tmp/manifests:Z" kubevirtci-tools/check_image_pull_policy /tmp/manifests

}

main "$@"