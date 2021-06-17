#!/bin/bash

set -euo pipefail

provision_dir="$1"
[ ! -d $provision_dir ] && echo "directory $provision_dir does not exist!" && exit 1
[ ! -d $provision_dir/manifests ] && echo "directory $provision_dir/manifests does not exist!" && exit 1

tag_name=$(curl -s -L -f https://api.github.com/repos/kubevirt/containerized-data-importer/releases/latest | jq -r '.tag_name')

if [ $(find "$provision_dir/manifests" -name "cdi-*.yaml" | wc -l) -gt 0 ]; then
    echo "$provision_dir/manifests/cdi-*.yaml already exists!"
    find "$provision_dir/manifests" -name "cdi-*.yaml" -print
    exit 1
fi

(
    for file in $(
        latest_release=$(curl -s -L -f https://api.github.com/repos/kubevirt/containerized-data-importer/releases/latest | jq -r '.url')
        curl -s -L -f $latest_release/assets | jq -r '.[] | select( .name | startswith( "cdi" ) ) | .browser_download_url'
    ); do
        echo '---'
        curl -s -L -f $file
    done
) > $provision_dir/manifests/cdi-${tag_name#v}.yaml
