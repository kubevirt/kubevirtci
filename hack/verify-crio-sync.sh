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

get_repo_type() {
    local provision_script=$1
    if grep -q "isv_cri-o_prerelease" "$provision_script"; then
        echo "prerelease"
    else
        echo "stable"
    fi
}

main() {
    if [ ! -d "$K8S_PROVIDERS_DIR" ]; then
        echo "ERROR: K8s providers directory not found: $K8S_PROVIDERS_DIR"
        exit 1
    fi

    local mismatches=0
    local total_providers=0

    for provider_dir in "$K8S_PROVIDERS_DIR"/*/; do
        [ -d "$provider_dir" ] || continue

        total_providers=$((total_providers + 1))

        local provider_name
        provider_name=$(basename "$provider_dir")
        local version_file="${provider_dir}version"
        local provision_script="${provider_dir}k8s_provision.sh"

        if [ ! -f "$version_file" ] || [ ! -f "$provision_script" ]; then
            echo "WARN: Skipping $provider_name: missing version or k8s_provision.sh"
            continue
        fi

        local k8s_version
        k8s_version=$(tr -d '\n' < "$version_file")
        local minor_version
        minor_version=$(get_minor_version "$k8s_version")
        local current_crio
        current_crio=$(get_current_crio_version "$provision_script")
        local repo_type
        repo_type=$(get_repo_type "$provision_script")
        local expected_crio="$minor_version"

        local expected_repo_type="stable"
        if is_prerelease_version "$k8s_version"; then
            expected_repo_type="prerelease"
        fi

        echo "Provider $provider_name: K8s=$k8s_version CRI-O=$current_crio($repo_type) expected=$expected_crio($expected_repo_type)"

        if [ "$current_crio" != "$expected_crio" ]; then
            echo "  FAIL: CRI-O version mismatch"
            mismatches=$((mismatches + 1))
        fi

        if [ "$repo_type" != "$expected_repo_type" ]; then
            echo "  FAIL: repository type mismatch"
            mismatches=$((mismatches + 1))
        fi
    done

    echo ""
    if [ "$mismatches" -eq 0 ]; then
        echo "OK: All $total_providers providers have correct CRI-O configurations"
        return 0
    else
        echo "FAIL: $mismatches issue(s) found across $total_providers providers"
        echo "Run: ./hack/sync-crio-versions.sh"
        return 1
    fi
}

main
