#!/bin/bash
set -xe

source ${KUBEVIRTCI_PATH}/cluster/kind/common.sh

MANIFESTS_DIR="${KUBEVIRTCI_PATH}/cluster/$KUBEVIRT_PROVIDER/manifests"
CERTCREATOR_PATH="${KUBEVIRTCI_PATH}/cluster/$KUBEVIRT_PROVIDER/certcreator"
KUBECONFIG_PATH="${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kubeconfig"

MASTER_NODE="${CLUSTER_NAME}-control-plane"
WORKER_NODE_ROOT="${CLUSTER_NAME}-worker"

OPERATOR_GIT_HASH=8d3c30de8ec5a9a0c9eeb84ea0aa16ba2395cd68  # release-4.4
SRIOV_OPERATOR_NAMESPACE="sriov-network-operator"

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

#move the pf to the node
mkdir -p /var/run/netns/
export pid="$(docker inspect -f '{{.State.Pid}}' $SRIOV_NODE)"
ln -sf /proc/$pid/ns/net "/var/run/netns/$SRIOV_NODE"

sriov_pfs=( /sys/class/net/*/device/sriov_numvfs )


for ifs in "${sriov_pfs[@]}"; do
  ifs_name="${ifs%%/device/*}"
  ifs_name="${ifs_name##*/}"

  if [ $(echo "${PF_BLACKLIST[@]}" | grep "${ifs_name}") ]; then
    continue
  fi

  # We set the variable below only in the first iteration as we need only one PF
  # to inject into the Network Configuration manifest. We need to move all pfs to
  # the node's namespace and for that reason we do not interrupt the loop.
  if [ -z "$NODE_PF" ]; then
    # These values are used to populate the network definition policy yaml.
    # We just use the first suitable pf
    # We need the num of vfs because if we don't set this value equals to the total, in case of mellanox
    # the sriov operator will trigger a node reboot to update the firmware
    export NODE_PF="$ifs_name"
    export NODE_PF_NUM_VFS=$(cat /sys/class/net/"$NODE_PF"/device/sriov_totalvfs)
  fi
  ip link set "$ifs_name" netns "$SRIOV_NODE"
done

SRIOV_NODE_CMD="docker exec -it -d ${SRIOV_NODE}"
${SRIOV_NODE_CMD} mount -o remount,rw /sys     # kind remounts it as readonly when it starts, we need it to be writeable
${SRIOV_NODE_CMD} chmod 666 /dev/vfio/vfio
_kubectl label node $SRIOV_NODE sriov=true
