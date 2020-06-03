#!/bin/bash -e
set -x

source ${KUBEVIRTCI_PATH}/cluster/kind/common.sh

MANIFESTS_DIR="${KUBEVIRTCI_PATH}/cluster/$KUBEVIRT_PROVIDER/manifests"
CSRCREATORPATH="${KUBEVIRTCI_PATH}/cluster/$KUBEVIRT_PROVIDER/csrcreator"
KUBECONFIG_PATH="${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kubeconfig"

MASTER_NODE="${CLUSTER_NAME}-control-plane"
WORKER_NODE_ROOT="${CLUSTER_NAME}-worker"

OPERATOR_GIT_HASH=8d3c30de8ec5a9a0c9eeb84ea0aa16ba2395cd68  # release-4.4

# This function gets a command string and invoke it
# until the command returns an empty string or until timeout
function retry {
  local -r tries=$1
  local -r wait_time=$2
  local -r action=$3
  local -r wait_message=$4

  local result=$(eval $action)
  for i in $(seq $tries); do
    if [[ -z $result ]] ; then
      echo "[$i/$tries] $wait_message"
      sleep $wait_time
      result=$(eval $action)
    else
      return 0
    fi
  done

  return 1
}

function wait_pod {
  local namespace=$1
  local label=$2

  local -r tries=60
  local -r wait_time=5

  local -r wait_message="Waiting for pods with $label to create"
  local -r error_message="Pods with  label $label at $namespace namespace found"

  if [[ $namespace != "" ]];then
    namespace="-n $namespace"
  fi

  if [[ $label != "" ]];then
    label="-l $label"
  fi

  local -r action="_kubectl get pod $namespace $label -o custom-columns=NAME:.metadata.name --no-headers"

  retry "$tries" "$wait_time" "$action" "$wait_message" && return  0
  echo $error_message && return 1
}

function wait_k8s_object {
  local -r object_type=$1
  local -r name=$2
  local namespace=$3

  local -r tries=60
  local -r wait_time=3

  local -r wait_message="Waiting for $object_type $name"
  local -r error_message="$object_type $name at $namespace namespace found"

  if [[ $namespace != "" ]];then
    namespace="-n $namespace"
  fi

  local -r action="_kubectl get $object_type $name $namespace -o custom-columns=NAME:.metadata.name --no-headers"

  retry "$tries" "$wait_time" "$action" "$wait_message" && return 0
  echo $error_message && return  1
}

# not using kubectl wait since with the sriov operator the pods get restarted a couple of times and this is
# more reliable
function wait_pods_ready {
    while [ -n "$(_kubectl get pods --all-namespaces -o'custom-columns=status:status.containerStatuses[*].ready,metadata:metadata.name' --no-headers | grep false)" ]; do
        echo "Waiting for all pods to become ready ..."
        _kubectl get pods --all-namespaces -o'custom-columns=status:status.containerStatuses[*].ready,metadata:metadata.name' --no-headers
        sleep 10
    done
}

function deploy_sriov_operator {
  operator_path=${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/sriov-network-operator-${OPERATOR_GIT_HASH}
  if [ ! -d $operator_path ]; then
    curl -L https://github.com/openshift/sriov-network-operator/archive/${OPERATOR_GIT_HASH}/sriov-network-operator.tar.gz | tar xz -C ${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/
  fi

  pushd $operator_path
    export RELEASE_VERSION=4.4
    export SRIOV_NETWORK_OPERATOR_IMAGE=quay.io/openshift/origin-sriov-network-operator:${RELEASE_VERSION}
    export SRIOV_NETWORK_CONFIG_DAEMON_IMAGE=quay.io/openshift/origin-sriov-network-config-daemon:${RELEASE_VERSION}
    export SRIOV_NETWORK_WEBHOOK_IMAGE=quay.io/openshift/origin-sriov-network-webhook:${RELEASE_VERSION}
    export NETWORK_RESOURCES_INJECTOR_IMAGE=quay.io/openshift/origin-sriov-dp-admission-controller:${RELEASE_VERSION}
    export SRIOV_CNI_IMAGE=quay.io/openshift/origin-sriov-cni:${RELEASE_VERSION}
    export SRIOV_DEVICE_PLUGIN_IMAGE=quay.io/openshift/origin-sriov-network-device-plugin:${RELEASE_VERSION}
    export OPERATOR_EXEC=${KUBECTL}
    make deploy-setup-k8s SHELL=/bin/bash  # on prow nodes the default shell is dash and some commands are not working
  popd

  pushd "${CSRCREATORPATH}"
    go run . -namespace sriov-network-operator -secret operator-webhook-service -hook operator-webhook -kubeconfig $KUBECONFIG_PATH
    go run . -namespace sriov-network-operator -secret network-resources-injector-secret -hook network-resources-injector -kubeconfig $KUBECONFIG_PATH
  popd
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

#move the pf to the node
mkdir -p /var/run/netns/
export pid="$(docker inspect -f '{{.State.Pid}}' $SRIOV_NODE)"
ln -sf /proc/$pid/ns/net "/var/run/netns/$SRIOV_NODE"

sriov_pfs=( /sys/class/net/*/device/sriov_numvfs )


for ifs in "${sriov_pfs[@]}"; do
  ifs_name="${ifs%%/device/*}"
  ifs_name="${ifs_name##*/}"

  if [ $(echo "${PF_BLACKLIST[@]}" | grep -q "${ifs_name}") ]; then
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


# deploy multus
_kubectl create -f $MANIFESTS_DIR/multus.yaml

# give them some time to create pods before checking pod status
sleep 10

# make sure all containers are ready
wait_pods_ready

SRIOV_NODE_CMD="docker exec -it -d ${SRIOV_NODE}"

${SRIOV_NODE_CMD} mount -o remount,rw /sys     # kind remounts it as readonly when it starts, we need it to be writeable

deploy_sriov_operator

_kubectl label node $SRIOV_NODE sriov=true

wait_pods_ready

# Ensure webook-configuration object created
wait_k8s_object "validatingwebhookconfiguration" "operator-webhook-config"  || exit 1
wait_k8s_object "mutatingwebhookconfiguration"   "operator-webhook-config"  || exit 1
wait_k8s_object "mutatingwebhookconfiguration"   "network-resources-injector-config"  || exit 1

_kubectl patch validatingwebhookconfiguration operator-webhook-config --patch '{"webhooks":[{"name":"operator-webhook.sriovnetwork.openshift.io", "clientConfig": { "caBundle": "'"$(cat $CSRCREATORPATH/operator-webhook.cert)"'" }}]}'
_kubectl patch mutatingwebhookconfiguration network-resources-injector-config --patch '{"webhooks":[{"name":"network-resources-injector-config.k8s.io", "clientConfig": { "caBundle": "'"$(cat $CSRCREATORPATH/network-resources-injector.cert)"'" }}]}'
_kubectl patch mutatingwebhookconfiguration operator-webhook-config --patch '{"webhooks":[{"name":"operator-webhook.sriovnetwork.openshift.io", "clientConfig": { "caBundle": "'"$(cat $CSRCREATORPATH/operator-webhook.cert)"'" }}]}'

# we need to sleep to wait for the configuration above the be picked up
sleep 60

# Substitute NODE_PF and NODE_PF_NUM_VFS then create SriovNetworkNodePolicy CR
envsubst < $MANIFESTS_DIR/network_config_policy.yaml | _kubectl create -f -

SRIOV_OPERATOR_NAMESPACE="sriov-network-operator"
SRIOV_CNI_LABEL="app=sriov-cni"
SRIOV_DEVICE_PLUGIN_LABEL="app=sriov-device-plugin"

# Ensure SriovNetworkNodePolicy CR is created
policy_name=$(cat $MANIFESTS_DIR/network_config_policy.yaml | grep 'name:' | awk '{print $2}')
wait_k8s_object "SriovNetworkNodePolicy" $policy_name $SRIOV_OPERATOR_NAMESPACE  || exit 1

# Wait for sriov-operator to reconcile SriovNodeNetworkPolicy
# and create cni and device-plugin pods
wait_pod $SRIOV_OPERATOR_NAMESPACE $SRIOV_CNI_LABEL  || exit 1
wait_pod $SRIOV_OPERATOR_NAMESPACE $SRIOV_DEVICE_PLUGIN_LABEL || exit 1

# Wait for cni and device-plugin pods to be ready
_kubectl wait pods -n $SRIOV_OPERATOR_NAMESPACE -l $SRIOV_CNI_LABEL           --for condition=Ready --timeout 10m
_kubectl wait pods -n $SRIOV_OPERATOR_NAMESPACE -l $SRIOV_DEVICE_PLUGIN_LABEL --for condition=Ready --timeout 10m

${SRIOV_NODE_CMD} chmod 666 /dev/vfio/vfio
