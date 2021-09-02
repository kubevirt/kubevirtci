#!/bin/bash

set -euo pipefail

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

[ ! -d $provision_dir ] && echo "directory $provision_dir does not exist!" && exit 1
[ ! -d $provision_dir/manifests ] && echo "directory $provision_dir/manifests does not exist!" && exit 1

tag_name=$(curl -s -L -f https://api.github.com/repos/kubevirt/containerized-data-importer/releases/latest | jq -r '.tag_name')

if [ "$(find "$provision_dir/manifests" -name "cdi-*.yaml" | wc -l)" -gt 0 ]; then
    find "$provision_dir/manifests" -name "cdi-*.yaml" -print

    if [ ! -n "$force" ]; then
        echo "$provision_dir/manifests/cdi-*.yaml already exists!"
        exit 1
    fi

    echo "Deleting old cdi manifests"
    find "$provision_dir/manifests" -name "cdi-*.yaml" -delete
fi

(
    for file in $(
        latest_release=$(curl -s -L -f https://api.github.com/repos/kubevirt/containerized-data-importer/releases/latest | jq -r '.url')
        curl -s -L -f $latest_release/assets | jq -r '.[] | select( .name | startswith( "cdi" ) ) | .browser_download_url'
    ); do
        echo '---'
        curl -s -L -f $file
    done
) >$provision_dir/manifests/cdi-${tag_name#v}.yaml
