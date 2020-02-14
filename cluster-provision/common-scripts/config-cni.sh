#!/bin/env bash

set -e

pod_cidr="10.244.0.0/16"

function configure_cni(){
    yum install -y patch
    cni_name=$1
    if ! ls $cni_name > /dev/null 2>&1; then
        return 1
    fi
    cni_scripts_dir=$cni_name/script.d
    if [[ -d $cni_scripts_dir ]]; then
        for cni_script in $(ls $cni_scripts_dir); do
            $cni_scripts_dir/$cni_script $cni_name
        done
    fi
}
export -f configure_cni
