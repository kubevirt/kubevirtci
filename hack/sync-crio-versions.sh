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
# Copyright The KubeVirt Authors.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K8S_PROVIDERS_DIR="${SCRIPT_DIR}/../cluster-provision/k8s"

get_minor_version() {
    local version=$1
    echo "$version" | sed -E 's/^([0-9]+\.[0-9]+).*/\1/'
}

get_current_crio_version() {
    local provision_script=$1
    grep -E '^export CRIO_VERSION=' "$provision_script" | sed 's/export CRIO_VERSION=//'
}

is_prerelease_version() {
    local version=$1
    [[ "$version" == *"alpha"* ]] || [[ "$version" == *"beta"* ]] || [[ "$version" == *"rc"* ]]
}

update_crio_version() {
    local provision_script=$1
    local new_version=$2
    local is_prerelease=$3

    sed -i.bak "s/^export CRIO_VERSION=.*/export CRIO_VERSION=${new_version}/" "$provision_script"

    if [ "$is_prerelease" = "true" ]; then
        sed -i.bak "s|\[isv_cri-o_stable_v\${CRIO_VERSION}\]|[isv_cri-o_prerelease_v\${CRIO_VERSION}]|" "$provision_script"
        sed -i.bak "s|name=CRI-O v\${CRIO_VERSION} (Stable) (rpm)|name=CRI-O v\${CRIO_VERSION} (Prerelease) (rpm)|" "$provision_script"
        sed -i.bak "s|baseurl=https://storage.googleapis.com/kubevirtci-crio-mirror/isv_cri-o_stable_v\${CRIO_VERSION}|baseurl=https://download.opensuse.org/repositories/isv:/cri-o:/prerelease:/v\${CRIO_VERSION}:/build/rpm|" "$provision_script"
    else
        sed -i.bak "s|\[isv_cri-o_prerelease_v\${CRIO_VERSION}\]|[isv_cri-o_stable_v\${CRIO_VERSION}]|" "$provision_script"
        sed -i.bak "s|name=CRI-O v\${CRIO_VERSION} (Prerelease) (rpm)|name=CRI-O v\${CRIO_VERSION} (Stable) (rpm)|" "$provision_script"
        sed -i.bak "s|baseurl=https://download.opensuse.org/repositories/isv:/cri-o:/prerelease:/v\${CRIO_VERSION}:/build/rpm|baseurl=https://storage.googleapis.com/kubevirtci-crio-mirror/isv_cri-o_stable_v\${CRIO_VERSION}|" "$provision_script"
    fi

    rm -f "${provision_script}.bak"
}

main() {
    if [ ! -d "$K8S_PROVIDERS_DIR" ]; then
        echo "ERROR: K8s providers directory not found: $K8S_PROVIDERS_DIR"
        exit 1
    fi

    local changes_made=0
    local total_providers=0
    local updated_files=()

    for provider_dir in "$K8S_PROVIDERS_DIR"/*/; do
        [ -d "$provider_dir" ] || continue

        total_providers=$((total_providers + 1))

        local provider_name
        provider_name=$(basename "$provider_dir")
        local version_file="${provider_dir}version"
        local provision_script="${provider_dir}k8s_provision.sh"

        if [ ! -f "$version_file" ] || [ ! -f "$provision_script" ]; then
            echo "Skipping $provider_name: missing version or k8s_provision.sh"
            continue
        fi

        local k8s_version
        k8s_version=$(tr -d '\n' < "$version_file")
        local minor_version
        minor_version=$(get_minor_version "$k8s_version")
        local current_crio
        current_crio=$(get_current_crio_version "$provision_script")
        local expected_crio="$minor_version"

        echo "Provider $provider_name: K8s=$k8s_version CRI-O=$current_crio expected=$expected_crio"

        if [ "$current_crio" != "$expected_crio" ]; then
            local prerelease="false"
            if is_prerelease_version "$k8s_version"; then
                prerelease="true"
            fi
            update_crio_version "$provision_script" "$expected_crio" "$prerelease"
            updated_files+=("cluster-provision/k8s/$provider_name/k8s_provision.sh")
            changes_made=$((changes_made + 1))
            echo "  Updated CRI-O $current_crio -> $expected_crio (prerelease=$prerelease)"
        fi
    done

    echo ""
    echo "Scanned $total_providers providers, updated $changes_made"

    if [ $changes_made -gt 0 ]; then
        echo "Changed files:"
        for f in "${updated_files[@]}"; do
            echo "  $f"
        done
    fi
}

main
