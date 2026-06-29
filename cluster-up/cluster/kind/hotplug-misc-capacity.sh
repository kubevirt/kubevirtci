#!/bin/bash
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

SCRIPT_PATH=$(dirname "$(realpath "$0")")
KUBEVIRTCI_PATH="$(realpath "${SCRIPT_PATH}/../../..")/"

source "${KUBEVIRTCI_PATH}/hack/detect_cri.sh"
export CRI_BIN=${CRI_BIN:-$(detect_cri)}

TARGET_PATH="/sys/fs/cgroup/misc.capacity"
HOTPLUG_BASE_DIR="${HOTPLUG_BASE_DIR:-/tmp/kubevirtci-sev-hotplug}"
DEFAULT_MISC_CAPACITY_CONTENT=$'sev_es 42\n'

CLUSTER_NAME="${CLUSTER_NAME:-}"
TARGET_NODE="${TARGET_NODE:-}"
MISC_CAPACITY_CONTENT="${MISC_CAPACITY_CONTENT:-$DEFAULT_MISC_CAPACITY_CONTENT}"
SOURCE_FILE=""
TEMP_SOURCE_FILE=""

usage() {
    cat <<EOF
Hotplug a readable misc.capacity file into running kind node containers.

This script uses a privileged overlay over /sys/fs/cgroup inside each node.
It is a runtime hack for experiments, not a persistent cluster configuration.

Usage:
  $(basename "$0") [--cluster-name <name>] [--node <node-name>] \
    [--source-file <path> | --content <text>]

Options:
  --cluster-name <name>  kind cluster name. If omitted, auto-detects exactly
                         one running kind cluster.
  --node <node-name>     only patch one node. Defaults to all nodes in the
                         selected cluster.
  --source-file <path>   read misc.capacity contents from a local file.
  --content <text>       inline misc.capacity contents.
                         Default: sev_es 42
  --help                 show this help text.

Environment overrides:
  CRI_BIN                podman or docker runtime.
  HOTPLUG_BASE_DIR       per-node scratch directory inside the node.
  TARGET_PATH            target path inside the node.
EOF
}

die() {
    echo "Error: $*" >&2
    exit 1
}

cleanup() {
    if [[ -n "${TEMP_SOURCE_FILE}" && -f "${TEMP_SOURCE_FILE}" ]]; then
        rm -f "${TEMP_SOURCE_FILE}"
    fi
}

trap cleanup EXIT

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
        --cluster-name)
            CLUSTER_NAME="${2:?missing value for --cluster-name}"
            shift 2
            ;;
        --node)
            TARGET_NODE="${2:?missing value for --node}"
            shift 2
            ;;
        --source-file)
            SOURCE_FILE="$(realpath "${2:?missing value for --source-file}")"
            shift 2
            ;;
        --content)
            MISC_CAPACITY_CONTENT="${2:?missing value for --content}"
            shift 2
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        *)
            die "unknown argument: $1"
            ;;
        esac
    done
}

resolve_cluster_name() {
    if [[ -n "${CLUSTER_NAME}" ]]; then
        return
    fi

    local clusters=()
    mapfile -t clusters < <(
        "${CRI_BIN}" ps \
            --filter "label=io.x-k8s.kind.cluster" \
            --format '{{.Label "io.x-k8s.kind.cluster"}}' | awk 'NF' | sort -u
    )

    case "${#clusters[@]}" in
    0)
        die "no running kind clusters found; pass --cluster-name"
        ;;
    1)
        CLUSTER_NAME="${clusters[0]}"
        ;;
    *)
        die "multiple running kind clusters found; pass --cluster-name"
        ;;
    esac
}

prepare_source_file() {
    if [[ -n "${SOURCE_FILE}" ]]; then
        [[ -f "${SOURCE_FILE}" ]] || die "source file not found: ${SOURCE_FILE}"
        return
    fi

    TEMP_SOURCE_FILE=$(mktemp)
    printf '%s' "${MISC_CAPACITY_CONTENT}" > "${TEMP_SOURCE_FILE}"
    SOURCE_FILE="${TEMP_SOURCE_FILE}"
}

get_nodes() {
    "${CRI_BIN}" ps \
        --filter "label=io.x-k8s.kind.cluster=${CLUSTER_NAME}" \
        --format '{{.Names}}' | awk 'NF'
}

ensure_target_is_under_cgroup_root() {
    [[ "${TARGET_PATH}" == /sys/fs/cgroup/* ]] || \
        die "TARGET_PATH must stay under /sys/fs/cgroup: ${TARGET_PATH}"
}

relative_target_path() {
    local target_path="${1}"
    echo "${target_path#/sys/fs/cgroup/}"
}

prepare_node() {
    local node="${1}"
    local base_dir="${2}"
    local lower_dir="${3}"
    local upper_dir="${4}"
    local work_dir="${5}"
    local merged_dir="${6}"
    local target_rel_path="${7}"

    "${CRI_BIN}" exec --privileged -i "${node}" bash -s -- \
        "${base_dir}" "${lower_dir}" "${upper_dir}" "${work_dir}" \
        "${merged_dir}" "${target_rel_path}" <<'EOF'
set -euo pipefail

base_dir="${1}"
lower_dir="${2}"
upper_dir="${3}"
work_dir="${4}"
merged_dir="${5}"
target_rel_path="${6}"
target_parent="$(dirname "${upper_dir}/${target_rel_path}")"

mkdir -p "${base_dir}" "${lower_dir}" "${upper_dir}" "${work_dir}" \
    "${merged_dir}" "${target_parent}"
mount --make-private /sys/fs/cgroup

if ! awk -v target="${lower_dir}" '$5 == target {found=1} END{exit !found}' \
    /proc/self/mountinfo; then
    mount --bind /sys/fs/cgroup "${lower_dir}"
fi
EOF
}

activate_node() {
    local node="${1}"
    local lower_dir="${2}"
    local upper_dir="${3}"
    local work_dir="${4}"
    local merged_dir="${5}"
    local target_path="${6}"

    "${CRI_BIN}" exec --privileged -i "${node}" bash -s -- \
        "${lower_dir}" "${upper_dir}" "${work_dir}" "${merged_dir}" \
        "${target_path}" <<'EOF'
set -euo pipefail

lower_dir="${1}"
upper_dir="${2}"
work_dir="${3}"
merged_dir="${4}"
target_path="${5}"

mkdir -p "${work_dir}" "${merged_dir}"

if ! awk -v target="${merged_dir}" '$5 == target {found=1} END{exit !found}' \
    /proc/self/mountinfo; then
    mount -t overlay overlay \
        -o "lowerdir=${lower_dir},upperdir=${upper_dir},workdir=${work_dir}" \
        "${merged_dir}"
fi

if ! awk '
    $5 == "/sys/fs/cgroup" && $0 ~ / - overlay overlay / {found=1}
    END {exit !found}
' /proc/self/mountinfo; then
    mount --bind "${merged_dir}" /sys/fs/cgroup
fi

test -r "${target_path}"
cat "${target_path}"
EOF
}

hotplug_node() {
    local node="${1}"
    local target_rel_path
    target_rel_path="$(relative_target_path "${TARGET_PATH}")"

    local node_base_dir="${HOTPLUG_BASE_DIR%/}"
    local node_lower_dir="${node_base_dir}/lower"
    local node_upper_dir="${node_base_dir}/upper"
    local node_work_dir="${node_base_dir}/work"
    local node_merged_dir="${node_base_dir}/merged"
    local node_target_path="${node_upper_dir}/${target_rel_path}"

    echo "[${node}] preparing hotplug directories"
    prepare_node "${node}" "${node_base_dir}" "${node_lower_dir}" \
        "${node_upper_dir}" "${node_work_dir}" "${node_merged_dir}" \
        "${target_rel_path}"

    echo "[${node}] writing $(basename "${SOURCE_FILE}") to ${node_target_path}"
    "${CRI_BIN}" exec -i "${node}" sh -c "cat > \"${node_target_path}\"" \
        < "${SOURCE_FILE}"

    echo "[${node}] activating overlay on ${TARGET_PATH}"
    activate_node "${node}" "${node_lower_dir}" "${node_upper_dir}" \
        "${node_work_dir}" "${node_merged_dir}" "${TARGET_PATH}"
}

main() {
    parse_args "$@"
    ensure_target_is_under_cgroup_root
    resolve_cluster_name
    prepare_source_file

    local nodes=()
    mapfile -t nodes < <(get_nodes)
    [[ "${#nodes[@]}" -gt 0 ]] || \
        die "no running nodes found for cluster: ${CLUSTER_NAME}"

    if [[ -n "${TARGET_NODE}" ]]; then
        local node_found=false
        local filtered_nodes=()
        for node in "${nodes[@]}"; do
            if [[ "${node}" == "${TARGET_NODE}" ]]; then
                filtered_nodes+=("${node}")
                node_found=true
                break
            fi
        done
        [[ "${node_found}" == true ]] || \
            die "node ${TARGET_NODE} is not part of cluster ${CLUSTER_NAME}"
        nodes=("${filtered_nodes[@]}")
    fi

    echo "Using runtime: ${CRI_BIN}"
    echo "Using cluster: ${CLUSTER_NAME}"
    echo "Target path: ${TARGET_PATH}"
    echo "Nodes: ${nodes[*]}"

    for node in "${nodes[@]}"; do
        hotplug_node "${node}"
    done

    echo "misc.capacity is now readable at ${TARGET_PATH}"
}

main "$@"
