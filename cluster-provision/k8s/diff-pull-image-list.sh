#!/bin/bash

set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function usage {
    cat <<EOF
Usage: $0 <k8s-cluster-dir>

    Compares the image list file with the result from calling fetch-images.sh on the cluster provision directory
    and manifests. Returns a list that shows the differences if any.

    Exits with non-zero exit code if the image file list exists and there are differences between the result of the call
    and the list content, exits with a zero exit code otherwise.

EOF
}

function check_args {
    if [ "$#" -lt 1 ]; then
        usage
        exit 1
    fi
    if [ ! -d "$DIR/$1" ]; then
        usage
        echo "Directory $DIR/$1 does not exist"
        exit 1
    fi
}

function main {
    check_args "$@"

    provision_dir="$DIR/$1"

    if [ ! -f "$provision_dir/pre-pull-images" ]; then
        exit 0
    fi

    new_list=$("$DIR/fetch-images.sh" "$1" "$1/pre-pull-images")
    if ! (echo "${new_list}") | diff -q "$provision_dir/pre-pull-images" - ; then
        echo "Image list check failed!"
        echo "Existing <-> generated:"
        (echo "${new_list}") | diff -y "$provision_dir/pre-pull-images" -
    else
        echo "Image list check succeeded."
    fi
}

main "$@"