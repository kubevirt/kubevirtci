#!/bin/bash

[ $(id -u) -ne 0 ] && echo "FATAL: this script requires sudo privileges" >&2 && exit 1

set -xe

PF_COUNT_PER_NODE=${PF_COUNT_PER_NODE:-1}
[ $PF_COUNT_PER_NODE -le 0 ] && echo "FATAL: PF_COUNT_PER_NODE must be a positive integer" >&2 && exit 1

SCRIPT_PATH=$(dirname "$(realpath "$0")")

source ${SCRIPT_PATH}/sriov-components/sriov_components.sh

CONFIGURE_VFS_SCRIPT_PATH="$SCRIPT_PATH/configure_vfs.sh"

SRIOV_COMPONENTS_NAMESPACE="sriov"
SRIOV_NODE_LABEL_KEY="sriov_capable"
SRIOV_NODE_LABEL_VALUE="true"
SRIOV_NODE_LABEL="$SRIOV_NODE_LABEL_KEY=$SRIOV_NODE_LABEL_VALUE"
SRIOVDP_RESOURCE_PREFIX="kubevirt.io"
SRIOVDP_RESOURCE_NAME="sriov_net"
VFS_DRIVER="vfio-pci"
VFS_DRIVER_KMODULE="vfio_pci"

function prepare_node_netns() {
  local -r node_name=$1
  local -r node_pid=$(docker inspect -f '{{.State.Pid}}' "$node_name")

  # Docker does not create the required symlink for a container netns
  # it perverts iplink from learning that container netns.
  # Thus it is necessary to create symlink between the current
  # worker node (container) netns to /var/run/netns (consumed by iplink)
  # Now the node container netns named with the node name will be visible.
  ln -sf "/proc/$node_pid/ns/net" "/var/run/netns/$node_name"
}

function move_pf_to_node_netns() {
  local -r node_name=$1
  local -r pf_name=$2

  # Move PF to node network-namespace
  ip link set "$pf_name" netns "$node_name"
  # Ensure current PF is up
  ip netns exec "$node_name" ip link set up dev "$pf_name"
  ip netns exec "$node_name" ip link show
}

function get_pfs_names() {
  local -r sriov_pfs=( $(find /sys/class/net/*/device/sriov_numvfs) )
  [ "${#sriov_pfs[@]}" -eq 0 ] && echo "FATAL: Could not find available sriov PFs on host" >&2 && return 1

  local pf_name
  local pf_names=()
  for pf in "${sriov_pfs[@]}"; do
    pf_name="${pf%%/device/*}"
    pf_name="${pf_name##*/}"
    if [ $(echo "${PF_BLACKLIST[@]}" | grep "${pf_name}") ]; then
      continue
    fi

    pfs_names+=( $pf_name )
  done

  echo "${pfs_names[@]}"
}

function configure_nodes_sriov_pfs_and_vfs() {
  local -r nodes_array=($1)
  local -r pfs_names_array=($2)
  local -r pf_count_per_node=$3

  local -r config_vf_script=$(basename "$CONFIGURE_VFS_SCRIPT_PATH")
  local pfs_to_move=()
  local pfs_array_offset=0
  local node_exec

  # 'iplink' learns which network namespaces there are by checking /var/run/netns
  mkdir -p /var/run/netns
  for node in "${nodes_array[@]}"; do
    prepare_node_netns "$node"

    ## Move PF's to node netns
    # Slice '$pfs_names_array' to have unique silce for each node
    # with '$pf_count_per_node' PF's names
    pfs_to_move=( "${pfs_names_array[@]:$pfs_array_offset:$pf_count_per_node}" )
    echo "Moving '${pfs_to_move[*]}' PF's to '$node' netns"
    for pf_name in "${pfs_to_move[@]}"; do
      move_pf_to_node_netns "$node" "$pf_name"
    done
    # Increment the offset for next slice
    pfs_array_offset=$((pfs_array_offset + pf_count_per_node))

    # KIND mounts sysfs as read-only by default, remount as R/W"
    node_exec="docker exec $node"
    $node_exec mount -o remount,rw /sys
    $node_exec chmod 666 /dev/vfio/vfio

    # Create and configure SRIOV Virtual Functions on SRIOV node
    docker cp "$CONFIGURE_VFS_SCRIPT_PATH" "$node:/"
    $node_exec bash -c "DRIVER=$VFS_DRIVER DRIVER_KMODULE=$VFS_DRIVER_KMODULE ./$config_vf_script"

    _kubectl label node $node $SRIOV_NODE_LABEL
  done
}

function validate_nodes_sriov_allocatable_resource() {
  local -r resource_name="$SRIOVDP_RESOURCE_PREFIX/$SRIOVDP_RESOURCE_NAME"
  local -r sriov_nodes=$(_kubectl get nodes -l $SRIOV_NODE_LABEL -o custom-columns=:.metadata.name --no-headers)

  local num_vfs
  for sriov_node in $sriov_nodes; do
    num_vfs=$(total_vfs_count_on_node "$sriov_node")
    sriov_components::wait_allocatable_resource "$sriov_node" "$resource_name" "$num_vfs"
  done
}

function total_vfs_count_on_node() {
  local -r node_name=$1
  local -r node_pid=$(docker inspect -f '{{.State.Pid}}' "$node_name")
  local -r pfs_sriov_numvfs=( $(cat /proc/$node_pid/root/sys/class/net/*/device/sriov_numvfs) )
  local total_vfs_on_node=0

  for num_vfs in "${pfs_sriov_numvfs[@]}"; do
    total_vfs_on_node=$((total_vfs_on_node + num_vfs))
  done

  echo "$total_vfs_on_node"
}

worker_nodes=($(_kubectl get nodes -l node-role.kubernetes.io/worker -o custom-columns=:.metadata.name --no-headers))
worker_nodes_count=${#worker_nodes[@]}
[ "$worker_nodes_count" -eq 0 ] && echo "FATAL: no worker nodes found" >&2 && exit 1

pfs_names=($(get_pfs_names))
pf_count="${#pfs_names[@]}"
[ "$pf_count" -eq 0 ] && echo "FATAL: Could not find available sriov PF's" >&2 && exit 1

total_pf_required=$((worker_nodes_count*PF_COUNT_PER_NODE))
[ "$pf_count" -lt "$total_pf_required" ] && \
  echo "FATAL: there are not enough PF's on the host, try to reduce PF_COUNT_PER_NODE
  Worker nodes count: $worker_nodes_count
  PF per node count:  $PF_COUNT_PER_NODE
  Total PF count required:  $total_pf_required" >&2 && exit 1

## Move SRIOV Physical Functions to worker nodes create VF's and configure their drivers
configure_nodes_sriov_pfs_and_vfs "${worker_nodes[*]}" "${pfs_names[*]}" "$PF_COUNT_PER_NODE"

## Deploy Multus and SRIOV components
sriov_components::deploy_multus
sriov_components::deploy \
  "${pfs_names[*]}" \
  "$VFS_DRIVER" \
  "$SRIOVDP_RESOURCE_PREFIX" "$SRIOVDP_RESOURCE_NAME" \
  "$SRIOV_NODE_LABEL_KEY" "$SRIOV_NODE_LABEL_VALUE"

# Verify that each sriov capable node has sriov VFs allocatable resource
validate_nodes_sriov_allocatable_resource
sriov_components::wait_pods_ready

_kubectl get nodes
_kubectl get pods -n $SRIOV_COMPONENTS_NAMESPACE
