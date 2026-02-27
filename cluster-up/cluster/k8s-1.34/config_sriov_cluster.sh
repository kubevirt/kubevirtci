#!/bin/bash

set -xe

SCRIPT_PATH=$(dirname "$(realpath "$0")")
CONFIGURE_VFS_SCRIPT_PATH="$SCRIPT_PATH/sriov-node/configure_vfs.sh"

# Set KUBEVIRTCI_PATH if not set
if [ -z "$KUBEVIRTCI_PATH" ]; then
    KUBEVIRTCI_PATH="$(cd "$(dirname "$0")/../../.." && pwd)"
fi

# Set KUBEVIRTCI_CONFIG_PATH if not set
if [ -z "$KUBEVIRTCI_CONFIG_PATH" ]; then
    KUBEVIRTCI_CONFIG_PATH="${HOME}/.kubevirtci"
fi

# Set KUBEVIRT_PROVIDER if not set
if [ -z "$KUBEVIRT_PROVIDER" ]; then
    export KUBEVIRT_PROVIDER="k8s-1.34"
fi

# Configuration
SRIOV_NODE_LABEL_KEY="sriov_capable"
SRIOV_NODE_LABEL_VALUE="true"
SRIOV_NODE_LABEL="$SRIOV_NODE_LABEL_KEY=$SRIOV_NODE_LABEL_VALUE"
SRIOVDP_RESOURCE_PREFIX="kubevirt.io"
SRIOVDP_RESOURCE_NAME="sriov_net"
VFS_DRIVER="vfio-pci"
VFS_DRIVER_KMODULE="vfio_pci"
VFS_COUNT="${VFS_COUNT:-6}"
KUBEVIRT_USE_DRA=${KUBEVIRT_USE_DRA:-false}

# Source SR-IOV components deployment functions
source "${SCRIPT_PATH}/sriov-components/sriov_components.sh"

# SSH function to execute commands on nodes
function ssh_to_node() {
  local node_name=$1
  shift
  "${KUBEVIRTCI_PATH}/cluster-up/ssh.sh" "$node_name" "$@"
}

# Configure VFs on a single node
function configure_node_vfs() {
  local node_name=$1
  
  echo "===== Configuring SR-IOV on $node_name ====="
  
  # Copy configure script to node using base64 encoding
  echo "Copying configure_vfs.sh to $node_name..."
  local encoded_script=$(base64 -w 0 "$CONFIGURE_VFS_SCRIPT_PATH")
  ssh_to_node "$node_name" "echo '$encoded_script' | base64 -d > /tmp/configure_vfs.sh && chmod +x /tmp/configure_vfs.sh"
  
  # Run configure script on node
  echo "Running VF configuration on $node_name..."
  ssh_to_node "$node_name" "sudo DRIVER=$VFS_DRIVER DRIVER_KMODULE=$VFS_DRIVER_KMODULE VFS_COUNT=$VFS_COUNT bash /tmp/configure_vfs.sh"
  
  # Label the node
  kubectl label node "$node_name" "$SRIOV_NODE_LABEL" --overwrite
  
  echo "===== SR-IOV configuration completed on $node_name ====="
}

# Get total VFs count on a node
function get_node_vfs_count() {
  local node_name=$1
  ssh_to_node "$node_name" "cat /sys/class/net/*/device/sriov_numvfs 2>/dev/null | awk '{s+=\$1} END {print s}'" 2>/dev/null || echo "0"
}

# Wait for allocatable resources to appear
function wait_allocatable_resource() {
  local node_name=$1
  local expected_value=$2
  
  sriov_components::wait_allocatable_resource "$node_name" "$SRIOVDP_RESOURCE_PREFIX/$SRIOVDP_RESOURCE_NAME" "$expected_value"
}

# Wait for allocatable resources to appear
function wait_allocatable_resource() {
  local node_name=$1
  local resource_name="$SRIOVDP_RESOURCE_PREFIX/$SRIOVDP_RESOURCE_NAME"
  local expected_value=$2
  
  echo "Waiting for $node_name to have $expected_value allocatable $resource_name resources..."
  
  local tries=48
  local wait_time=10
  
  for i in $(seq 1 $tries); do
    local current=$(kubectl get node "$node_name" -o jsonpath="{.status.allocatable.kubevirt\.io\/sriov_net}" 2>/dev/null || echo "0")
    
    if [ "$current" == "$expected_value" ]; then
      echo "✓ Node $node_name has $expected_value allocatable $resource_name resources"
      return 0
    fi
    
    echo "[$i/$tries] Current: $current, Expected: $expected_value, waiting ${wait_time}s..."
    sleep $wait_time
  done
  
  echo "ERROR: Timeout waiting for allocatable resources on $node_name"
  return 1
}

# Main execution
echo "===== Starting SR-IOV Cluster Configuration ====="

# Get worker nodes
worker_nodes=$(kubectl get nodes -l node-role.kubernetes.io/worker -o custom-columns=:.metadata.name --no-headers)
worker_nodes_array=($worker_nodes)
worker_nodes_count=${#worker_nodes_array[@]}

if [ "$worker_nodes_count" -eq 0 ]; then
  echo "FATAL: no worker nodes found" >&2
  exit 1
fi

echo "Found $worker_nodes_count worker node(s): ${worker_nodes_array[*]}"

# Configure VFs on each worker node
for node in "${worker_nodes_array[@]}"; do
  configure_node_vfs "$node"
done

echo "===== SR-IOV VF Configuration Complete ====="
echo ""
echo "Node SR-IOV Status:"
for node in "${worker_nodes_array[@]}"; do
  vf_count=$(get_node_vfs_count "$node")
  echo "  $node: $vf_count VFs configured"
done

echo ""
echo "===== Deploying Multus CNI ====="
sriov_components::deploy_multus

# Collect PF names from all nodes
PFS_IN_USE="eth1"  # Our SR-IOV interface

if [[ "$KUBEVIRT_USE_DRA" != "true" ]]; then
  echo ""
  echo "===== Deploying SR-IOV Device Plugin ====="
  sriov_components::deploy \
    "$PFS_IN_USE" \
    "$VFS_DRIVER" \
    "$SRIOVDP_RESOURCE_PREFIX" "$SRIOVDP_RESOURCE_NAME" \
    "$SRIOV_NODE_LABEL_KEY" "$SRIOV_NODE_LABEL_VALUE"

  echo ""
  echo "===== Waiting for Allocatable Resources ====="
  # Verify that each sriov capable node has sriov VFs allocatable resource
  for node in "${worker_nodes_array[@]}"; do
    vf_count=$(get_node_vfs_count "$node")
    if [ "$vf_count" -gt 0 ]; then
      wait_allocatable_resource "$node" "$vf_count"
    fi
  done
else
  echo ""
  echo "===== Deploying SR-IOV DRA Driver ====="
  sriov_components::deploy_dra
fi

echo ""
echo "===== Waiting for All Pods to be Ready ====="
sriov_components::wait_pods_ready

echo ""
echo "===== SR-IOV Cluster Configuration Complete ====="
kubectl get nodes -l "$SRIOV_NODE_LABEL"
kubectl get pods -n sriov 2>/dev/null || echo "SR-IOV namespace not found (expected if using DRA)"

echo ""
echo "To verify VFs on a node, run:"
echo "  ./cluster-up/ssh.sh node01"
echo "  lspci | grep -i virtual"
echo ""
