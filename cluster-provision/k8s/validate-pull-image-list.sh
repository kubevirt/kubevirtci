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

    non_conformant_images=$(mktemp)
    trap "rm -f ${non_conformant_images}" EXIT SIGINT SIGTERM

    while read image; do
        if [[ ! "${image}" =~ ^([a-z0-9\_\.]+[/-]?)+:[a-z0-9\_\.\-]+$ ]]; then
            echo "${image} - failed expression check" >> "${non_conformant_images}"
        fi
        if [[ "${image}" =~ latest$ ]]; then
            echo "${image} - latest not allowed" >> "${non_conformant_images}"
        fi
    done

    if [ ! -s "${non_conformant_images}" ]; then
        exit 0
    fi

    echo "Image check failed!"
    cat "${non_conformant_images}"

}

main "$@"