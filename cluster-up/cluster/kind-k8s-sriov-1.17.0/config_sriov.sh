#!/bin/bash

set -ex

if [ "$(id -u)" -ne 0 ]; then 
  echo "This script requires sudo privileges"
  exit 1
fi

source cluster-up/hack/common.sh 
source ${KUBEVIRTCI_PATH}/cluster/kind/common.sh

PF_BLACKLIST=${PF_BLACKLIST:-none}
SRIOV_PF_DEVICES_ROOT_DIR=( /sys/class/net/*/device/sriov_numvfs )
SRIOV_WORKER_NODES_LABEL="sriov=true"

SRIOV_POLICY_FILE_PATH=${SRIOV_POLICY_FILE_PATH:-none}
PATCHED_SRIOV_POLICY_FILE_PATH=${PATCHED_SRIOV_POLICY_FILE_PATH:-node}

echo "Geting SRIOV PF interfaces names"
pf_names=()
pf_vfs_count=()
for interface in "${SRIOV_PF_DEVICES_ROOT_DIR[@]}"; do
  interface_name="${interface%%/device/*}"
  interface_name="${interface_name##*/}"

  if [ $(echo "${PF_BLACKLIST[@]}" | grep "${interface_name}") ]; then
    continue
  fi

  pf_names+=( $interface_name )
done
echo "${pf_names[@]}"

echo "Getting total VF's count that supported by the card"
vfs_count=$(cat "/sys/class/net/${pf_names[0]}/device/sriov_totalvfs")
echo $vfs_count

echo "Patch PF's names and VF's count to SriovNetworkNodePolicy"
cp -f $SRIOV_POLICY_FILE_PATH $PATCHED_SRIOV_POLICY_FILE_PATH
sed -i "s?numVfs:?numVfs: $vfs_count?g" $PATCHED_SRIOV_POLICY_FILE_PATH
pf_names_comma_separeted=$(echo "${pf_names[@]}" | sed  's/ /\,/g')
sed -i "s?pfNames:?pfNames: [$pf_names_comma_separeted]?g" $PATCHED_SRIOV_POLICY_FILE_PATH
cat $PATCHED_SRIOV_POLICY_FILE_PATH

echo "Get worker nodes"
worker_nodes=( $(_kubectl get nodes -l node-role.kubernetes.io/worker -o custom-columns=:.metadata.name --no-headers) )
echo "${worker_nodes[@]}"

echo "Attach PF's to workers nodes"
mkdir -p /var/run/netns/

pf_count="${#pf_names[@]}"
for i in $(seq $pf_count); do 
  index=$((i-1))
  current_node="${worker_nodes[$index]}"

  if [ -z $current_node ]; then
    echo "All workers were configured"
    break
  fi

  current_pf="${pf_names[$index]}"
  if [ -z $current_pf ]; then 
    echo "All PF's were attached to worker nodes"
    break
  fi

  echo "[$current_node] Attaching PF '$current_pf' to node network namespace"
  pid="$(docker inspect -f '{{.State.Pid}}' $current_node)"
  current_node_network_namespace=$current_node

  # Create symlink to current worker node (container) network-namespace
  # at /var/run/netns (consumned by iplink) so it will be visibale by iplink.
  # This is necessary since docker does not creating the requierd 
  # symlink for a container network-namesapce.
  ln -sf /proc/$pid/ns/net "/var/run/netns/$current_node_network_namespace"

  # Move current PF to current node network-namespace
  ip link set $current_pf netns $current_node_network_namespace

  # Ensure current PF is up
  ip netns exec $current_node_network_namespace ip link set up dev $current_pf

  echo "node '$current_node' network-namespace '$current_node_network_namespace' links state"
  ip netns exec $current_node_network_namespace ip link show

  current_node_cmd="docker exec -it -d $current_node"
  # kind remounts it as readonly when it starts, we need it to be writeable
  ${current_node_cmd} mount -o remount,rw /sys

  # Ensure vfio binary is executable
  ${current_node_cmd} chmod 666 /dev/vfio/vfio

  echo "label node '$current_node' as sriov capable node '$SRIOV_WORKER_NODES_LABEL'"
  _kubectl label node $current_node $SRIOV_WORKER_NODES_LABEL
done
