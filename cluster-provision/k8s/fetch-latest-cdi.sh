#!/bin/bash

set -euo pipefail

function usage() {
    cat <<EOF
usage: $0 [-f] [<provision_dir>]

        Downloads the latest containerized-data-importer manifests from the github release page and stores them into the manifests directory of the provision_dir.

        If the provision_dir is omitted, it is assumed that the latest provision_dir should be the target

options:
    -f
        forces the download of the manifests, existing manifests are deleted beforehand
EOF
}

force=
while getopts ":f" opt; do
    case "${opt}" in
    f)
        force=true
        shift
        ;;
    \?)
        usage
        exit 1
        ;;
    esac
done

if [ "$#" -gt 0 ]; then
    provision_dir="$1"
else
    provision_dir=$(find "$(readlink --canonicalize "$(dirname "$0")")" -mindepth 1 -maxdepth 1 -type d -regex '^.*[0-9]\.[0-9]+$' -regextype 'posix-extended' | sort -rV | head -1)
fi

[ ! -d "$provision_dir" ] && echo "directory $provision_dir does not exist!" && exit 1
[ ! -d "$provision_dir/manifests" ] && echo "directory $provision_dir/manifests does not exist!" && exit 1

tag_name=$(curl -s -L -f https://api.github.com/repos/kubevirt/containerized-data-importer/releases/latest | jq -r '.tag_name')

if [ "$(find "$provision_dir/manifests" -name "cdi-*.yaml" | wc -l)" -gt 0 ]; then
    find "$provision_dir/manifests" -name "cdi-*.yaml" -print

    if [ -z "$force" ]; then
        echo "$provision_dir/manifests/cdi-*.yaml already exists!"
        exit 1
    fi

    echo "Deleting old cdi manifests"
    find "$provision_dir/manifests" -name "cdi-*.yaml" -delete
fi

cdi_release_json="$(pwd)/$(mktemp cdi_release_XXXXXXXX.json)"
curl -o "$cdi_release_json" -s -L -f https://api.github.com/repos/kubevirt/containerized-data-importer/releases/latest
function cleanup() {
    rm -f "$cdi_release_json"
}
trap cleanup EXIT SIGTERM
tag_name=$(jq -r '.tag_name' "$cdi_release_json")
cdi_version="${tag_name/#v/}"
(
    cd "$provision_dir/manifests"
    for file in $(
        jq -r '.assets[] | select( .name | startswith( "cdi" ) ) | .browser_download_url' "$cdi_release_json"
    ); do
        curl -O -s -L -f "$file"
    done

    while IFS= read -r -d '' file; do
        mv "$file" "${file/#\.\/cdi-/\.\/cdi-$cdi_version-}"
    done < <(find . -type f -name 'cdi*.yaml' -print0)
)
