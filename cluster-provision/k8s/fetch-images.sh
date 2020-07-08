#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function usage {
    cat <<EOF
Usage: $0 <k8s-cluster-dir>

    fetches all images from the cluster provision source that need to get pre pulled

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
    image_regex='([a-z0-9\_\.]+[/:-]?)+'
    image_regex_w_double_quotes='"?'"${image_regex}"'"?'
    (
        find "$provision_dir" -type f -name '*.sh' -print0 | \
            xargs -0 grep -iE '(docker|podman)[ _]pull[^ ]+ '"${image_regex_w_double_quotes}"
        find "$provision_dir" -type f -name '*.yaml' -print0 | \
            xargs -0 grep -iE 'image: '"${image_regex_w_double_quotes}"
    ) | grep -ioE "${image_regex_w_double_quotes}"'$' | sed -E 's/"//g' | sort | uniq
}

main "$@"