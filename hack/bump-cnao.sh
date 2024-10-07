#!/usr/bin/env bash
#
# This file is part of the KubeVirt project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Copyright 2021 Red Hat, Inc.

set -e

CNAO_RELEASES="https://github.com/kubevirt/cluster-network-addons-operator/releases/download"

# syntax:
# ./hack/bump-cnao.sh <CNAO_VERSION>

# usage example
# ./hack/bump-cnao.sh v0.54.0

function main() {
    cnao_version="${1:?cnao version not set or empty}"

    declare -a manifests_url
    manifests_url+=("${CNAO_RELEASES}/${cnao_version}/namespace.yaml")
    manifests_url+=("${CNAO_RELEASES}/${cnao_version}/network-addons-config.crd.yaml")
    manifests_url+=("${CNAO_RELEASES}/${cnao_version}/operator.yaml")
    manifests_url+=("${CNAO_RELEASES}/${cnao_version}/network-addons-config-example.cr.yaml")

    declare -a manifests
    for url in "${manifests_url[@]}"; do
        file="${url##*/}"
        expected_file_name=$file
        if [[ $file == "network-addons-config.crd.yaml" ]]; then
            expected_file_name="crd.yaml"
        fi
        if ! ls "./cluster-provision/gocli/opts/cnao/manifests/${expected_file_name}" > /dev/null; then
            echo "${expected_file_name} not found at kubevirtci folder"
            exit 1
        fi

        manifest=$(curl -Ls "${url}")
        if [[ "${manifest}" == "Not Found" ]]; then
            echo "${url} not found"
            exit 1
        fi

        manifests+=("${manifest}")
    done

    for i in "${!manifests[@]}"; do
        file="${manifests_url[i]##*/}"
        target=$file
        if [[ $file == "network-addons-config.crd.yaml" ]]; then
            target="crd.yaml"
        fi
        echo "${manifests[$i]}" > "./cluster-provision/gocli/opts/cnao/manifests/${target}"

        if [[ $target == "network-addons-config-example.cr.yaml" ]]; then
            sed -i '/ovs:/d' ./cluster-provision/gocli/opts/cnao/manifests/${target}
            sed -i '/kubevirtIpamController:/d' ./cluster-provision/gocli/opts/cnao/manifests/${target}
        fi
    done

    echo "cnao, provision: Bump CNAO to ${cnao_version}"
}

main "$@"
