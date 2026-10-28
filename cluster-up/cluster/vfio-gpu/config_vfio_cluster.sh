#!/usr/bin/env bash

set -e
set -o pipefail

SCRIPT_PATH="$(dirname "$(realpath "$0")")"
VFIO_NODE_WORKDIR="/tmp/fake-vfio"
VFIO_GPU_SRC=""

: "${KUBEVIRTCI_CONFIG_PATH:?FATAL: missing KUBEVIRTCI_CONFIG_PATH}"
: "${KUBEVIRT_PROVIDER:?FATAL: missing KUBEVIRT_PROVIDER}"
: "${FAKE_PCI_DEVICES:=8}"
: "${FAKE_IOMMU:=true}"

KUBECONFIG="${KUBEVIRTCI_CONFIG_PATH}/${KUBEVIRT_PROVIDER}/.kubeconfig"
KUBECTL="${KUBEVIRTCI_CONFIG_PATH}/${KUBEVIRT_PROVIDER}/.kubectl --kubeconfig=${KUBECONFIG}"

VFIO_NODE_LABEL_KEY="fake-vfio-capable"
VFIO_NODE_LABEL_VALUE="true"
VFIO_NODE_LABEL="${VFIO_NODE_LABEL_KEY}=${VFIO_NODE_LABEL_VALUE}"

source "${SCRIPT_PATH}/install_dra_example_driver.sh"

_kubectl() {
    ${KUBECTL} "$@"
}

ssh_to_node() {
    local node_name=$1
    shift
    "${SCRIPT_PATH}/../../ssh.sh" "$node_name" "$@"
}

encode_vfio_assets() {
    tar -C "${VFIO_GPU_SRC}" -cz fake-iommu fake-pci setup-fake-pci-host.sh setup-node-vfio.sh | base64 | tr -d '\n'
}

# Some fake VFIO tests require at least 8G of memory. Fail early if the cluster was started with less.
validate_memory() {
    local mem_value mem_unit mem_in_mb
    mem_value=$(echo "${KUBEVIRT_MEMORY_SIZE:-0}" | sed 's/[^0-9]//g')
    mem_unit=$(echo "${KUBEVIRT_MEMORY_SIZE:-0}" | sed 's/[^a-zA-Z]//g' | tr '[:lower:]' '[:upper:]')
    case "${mem_unit}" in
        G) mem_in_mb=$((mem_value * 1024)) ;;
        *) mem_in_mb=${mem_value} ;;
    esac
    if [ "${mem_in_mb}" -lt 8192 ]; then
        echo "FATAL: fake VFIO requires KUBEVIRT_MEMORY_SIZE >= 8G (8192M). Current value: ${KUBEVIRT_MEMORY_SIZE:-unset}" >&2
        exit 1
    fi
}

# Some fake VFIO tests require at least 8G of memory. Fail early if the cluster was started with less.
validate_memory() {
    local mem_value mem_unit mem_in_mb
    mem_value=$(echo "${KUBEVIRT_MEMORY_SIZE:-0}" | sed 's/[^0-9]//g')
    mem_unit=$(echo "${KUBEVIRT_MEMORY_SIZE:-0}" | sed 's/[^a-zA-Z]//g' | tr '[:lower:]' '[:upper:]')
    case "${mem_unit}" in
        G) mem_in_mb=$((mem_value * 1024)) ;;
        *) mem_in_mb=${mem_value} ;;
    esac
    if [ "${mem_in_mb}" -lt 8192 ]; then
        echo "FATAL: fake VFIO requires KUBEVIRT_MEMORY_SIZE >= 8G (8192M). Current value: ${KUBEVIRT_MEMORY_SIZE:-unset}" >&2
        exit 1
    fi
}

validate_inputs() {
    validate_memory
    clone_dra_example_driver
    VFIO_GPU_SRC="${DRA_EXAMPLE_DRIVER_DIR}/demo/vfio-gpu"
    if [ ! -d "${VFIO_GPU_SRC}/fake-iommu" ] || [ ! -d "${VFIO_GPU_SRC}/fake-pci" ]; then
        echo "FATAL: fake VFIO module sources not found under ${VFIO_GPU_SRC}" >&2
        echo "The dra-example-driver checkout must contain demo/vfio-gpu." >&2
        exit 1
    fi
    if [ ! -f "${VFIO_GPU_SRC}/setup-fake-pci-host.sh" ] || [ ! -f "${VFIO_GPU_SRC}/setup-node-vfio.sh" ]; then
        echo "FATAL: vfio-gpu setup scripts not found under ${VFIO_GPU_SRC}" >&2
        exit 1
    fi
}

copy_vfio_payload_to_node() {
    local node_name=$1
    local encoded_assets

    encoded_assets=$(encode_vfio_assets)

    echo "Copying fake VFIO payload from ${VFIO_GPU_SRC} to ${node_name}..."
    ssh_to_node "${node_name}" "rm -rf '${VFIO_NODE_WORKDIR}' && mkdir -p '${VFIO_NODE_WORKDIR}'"
    ssh_to_node "${node_name}" "printf '%s' '${encoded_assets}' | base64 -d | tar -xz -C '${VFIO_NODE_WORKDIR}'"
    ssh_to_node "${node_name}" "chmod +x '${VFIO_NODE_WORKDIR}/setup-node-vfio.sh' '${VFIO_NODE_WORKDIR}/setup-fake-pci-host.sh'"
}

run_node_vfio_setup() {
    local node_name=$1
    ssh_to_node "${node_name}" "sudo \
        FAKE_PCI_DEVICES='${FAKE_PCI_DEVICES}' \
        FAKE_PCI_DOMAIN='${FAKE_PCI_DOMAIN:-}' \
        FAKE_PCI_VENDOR_ID='${FAKE_PCI_VENDOR_ID:-}' \
        FAKE_PCI_DEVICE_ID='${FAKE_PCI_DEVICE_ID:-}' \
        FAKE_IOMMU='${FAKE_IOMMU}' \
        bash '${VFIO_NODE_WORKDIR}/setup-node-vfio.sh'"
}

configure_node_vfio() {
    local node_name=$1

    echo "===== Configuring fake VFIO on ${node_name} ====="
    copy_vfio_payload_to_node "${node_name}"
    run_node_vfio_setup "${node_name}"

    _kubectl label node "${node_name}" "${VFIO_NODE_LABEL}" --overwrite
    echo "===== fake VFIO configuration completed on ${node_name} ====="
}

print_node_status() {
    local node_name=$1
    local status_cmd='echo "  modules:"; lsmod | grep -E "^(fake_iommu|fake_pci|vfio_pci)[[:space:]]" | sed "s/^/    /" || true; echo "  vfio-pci devices:"; found=0; for d in /sys/bus/pci/drivers/vfio-pci/*:*; do [ -e "$d" ] || continue; found=1; echo "    $(basename "$d")"; done; [ "$found" -eq 1 ] || echo "    (none)"; echo "  /dev/vfio entries:"; ls -1 /dev/vfio 2>/dev/null | sed "s/^/    /" || echo "    (none)"'

    echo ""
    echo "Node fake VFIO status for ${node_name}:"
    ssh_to_node "${node_name}" "sudo bash -c $(printf "%q" "${status_cmd}")"
}

main() {
    validate_inputs

    echo "===== Starting fake VFIO cluster configuration ====="
    worker_nodes=$(_kubectl get nodes -l node-role.kubernetes.io/worker -o custom-columns=:.metadata.name --no-headers)
    worker_nodes_array=($worker_nodes)
    worker_nodes_count=${#worker_nodes_array[@]}

    if [ "${worker_nodes_count}" -eq 0 ]; then
        echo "FATAL: no worker nodes found" >&2
        exit 1
    fi

    echo "Found ${worker_nodes_count} worker node(s): ${worker_nodes_array[*]}"

    local node
    for node in "${worker_nodes_array[@]}"; do
        configure_node_vfio "${node}"
    done

    echo ""
    echo "===== Installing vfio-gpu DRA example driver ====="
    cluster::install_dra_example_driver

    echo ""
    echo "===== Waiting for DRA driver pods to be ready ====="
    _kubectl wait pods -n dra-example-driver --all --for=condition=Ready --timeout=300s

    for node in "${worker_nodes_array[@]}"; do
        print_node_status "${node}"
    done

    echo ""
    echo "===== fake VFIO cluster configuration complete ====="
    _kubectl get nodes -l "${VFIO_NODE_LABEL}"
    _kubectl get pods -n dra-example-driver
}

main "$@"
