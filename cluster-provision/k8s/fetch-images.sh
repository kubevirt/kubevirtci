#!/bin/bash

set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function usage {
    cat <<EOF
Usage: $0 <k8s-cluster-dir> [source-image-list]

    Fetches all images from the cluster provision source and manifests. Returns a list that is sorted and
    without double entries.

    If source-image-list is provided this is taken as an input and added to the result.

EOF
}

function check_args {
    if [ "$#" -lt 1 ]; then
        usage
        exit 1
    fi
    if [ ! -d "$1" ]; then
        usage
        echo "Directory $1 does not exist"
        exit 1
    fi
}

function main {
    check_args "$@"

    temp_file=$(mktemp)
    trap 'rm -f "${temp_file}"' EXIT SIGINT SIGTERM

    provision_dir="$1"
    image_regex='([a-z0-9\_\.]+[/-]?)+(@sha256)?:[a-z0-9\_\.\-]+'
    image_regex_w_double_quotes='"?'"${image_regex}"'"?'

    (
        # Avoid bailing out because of nothing found in scripts part
        set +e
        find "$provision_dir" -type f -name '*.sh' -print0 | \
            xargs -0 grep -iE '(docker|podman)[ _]pull[^ ]+ '"${image_regex_w_double_quotes}"
        find "$provision_dir" -type f -name '*.yaml' -print0 | \
            xargs -0 grep -iE '(image|value): '"${image_regex_w_double_quotes}"
        set -e
        # last `grep -v` is necessary to avoid trying to pre pull istio "images", as the regex also matches on values
        # from the generated istio deployment manifest
    ) | grep -ioE "${image_regex_w_double_quotes}"'$' | grep -v '.svc:' >> "${temp_file}"

    sed -E 's/"//g' "${temp_file}" | sort | uniq
}

main "$@"
