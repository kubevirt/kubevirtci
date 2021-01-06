#!/bin/bash
set -xe

source ${KUBEVIRTCI_PATH}/cluster/kind/common.sh

MASTER_NODE="${CLUSTER_NAME}-control-plane"
WORKER_NODE_ROOT="${CLUSTER_NAME}-worker"
PF_COUNT_PER_NODE=${PF_COUNT_PER_NODE:-1}

[ $PF_COUNT_PER_NODE -le 0 ] && echo "FATAL: PF_COUNT_PER_NODE must be a positive integer" >&2 && exit 1

SRIOV_NODE_LABEL=${SRIOV_NODE_LABEL:-"sriov=true"}
NODE_PFS_ANNOTATION=${NODE_PFS_ANNOTATION:-"node_pfs"}

function move_sriov_pfs_netns_to_node {
  local -r node=$1
  local -r pf_count_per_node=$2
  local -r pid="$(docker inspect -f '{{.State.Pid}}' $node)"
  local pf_array=()

  mkdir -p /var/run/netns/
  ln -sf /proc/$pid/ns/net "/var/run/netns/$node"

  local -r sriov_pfs=( $(find /sys/class/net/*/device/sriov_numvfs) )
  [ "${#sriov_pfs[@]}" -eq 0 ] && echo "FATAL: Could not find available sriov PFs" >&2 && return 1

  for pf in "${sriov_pfs[@]}"; do
    local pf_name="${pf%%/device/*}"
    pf_name="${pf_name##*/}"

    if [ $(echo "${PF_BLACKLIST[@]}" | grep "${pf_name}") ]; then
      continue
    fi

    ip link set "$pf_name" netns "$node"
    pf_array+=("$pf_name")
    [ "${#pf_array[@]}" -eq "$pf_count_per_node" ] && break
  done

  [ "${#pf_array[@]}" -lt "$pf_count_per_node" ] && \
    echo "FATAL: Not enough PFs allocated, PF_BLACKLIST (${PF_BLACKLIST[@]}), PF_COUNT_PER_NODE ${PF_COUNT_PER_NODE}" >&2 && \
    return 1

  echo "${pf_array[@]}"
}

# The first worker needs to be handled specially as it has no ending number, and sort will not work
# We add the 0 to it and we remove it if it's the candidate worker
WORKER=$(_kubectl get nodes | grep $WORKER_NODE_ROOT | sed "s/\b$WORKER_NODE_ROOT\b/${WORKER_NODE_ROOT}0/g" | sort -r | awk 'NR==1 {print $1}')
if [[ -z "$WORKER" ]]; then
  SRIOV_NODE=$MASTER_NODE
else
  SRIOV_NODE=$WORKER
fi

# this is to remove the ending 0 in case the candidate worker is the first one
if [[ "$SRIOV_NODE" == "${WORKER_NODE_ROOT}0" ]]; then
  SRIOV_NODE=${WORKER_NODE_ROOT}
fi

NODE_PFS=($(move_sriov_pfs_netns_to_node "$SRIOV_NODE" "$PF_COUNT_PER_NODE"))

SRIOV_NODE_CMD="docker exec -it -d ${SRIOV_NODE}"
${SRIOV_NODE_CMD} mount -o remount,rw /sys     # kind remounts it as readonly when it starts, we need it to be writeable
${SRIOV_NODE_CMD} chmod 666 /dev/vfio/vfio
_kubectl label node $SRIOV_NODE $SRIOV_NODE_LABEL

for pf in "${NODE_PFS[@]}"; do
  docker exec $SRIOV_NODE bash -c "echo 0 > /sys/class/net/$pf/device/sriov_numvfs"
done

used_pfs_yaml_format_array=$(sed 's/ /, /g' <<< "[\"${NODE_PFS[@]}\"]")
_kubectl annotate node $SRIOV_NODE "$NODE_PFS_ANNOTATION=$used_pfs_yaml_format_array"
