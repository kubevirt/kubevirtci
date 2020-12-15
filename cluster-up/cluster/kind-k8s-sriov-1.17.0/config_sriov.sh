#!/bin/bash
set -xe

source ${KUBEVIRTCI_PATH}/cluster/kind/common.sh

MANIFESTS_DIR="${KUBEVIRTCI_PATH}/cluster/$KUBEVIRT_PROVIDER/manifests"
CERTCREATOR_PATH="${KUBEVIRTCI_PATH}/cluster/$KUBEVIRT_PROVIDER/certcreator"
KUBECONFIG_PATH="${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kubeconfig"

MASTER_NODE="${CLUSTER_NAME}-control-plane"
WORKER_NODE_ROOT="${CLUSTER_NAME}-worker"

SRIOV_OPERATOR_NAMESPACE="sriov-network-operator"
NUM_PF_REQUIRED=${NUM_PF_REQUIRED:-1}

export RELEASE_VERSION=4.8.0
OPERATOR_GIT_HASH=49045c36efb9136813f049b9977fe2b93c0a46c0

re='^[0-9]+$'
if ! [[ $NUM_PF_REQUIRED =~ $re ]] || [[ $NUM_PF_REQUIRED -eq "0" ]]; then
  echo "FATAL: Wrong value of NUM_PF_REQUIRED, must be numeric, non zero, less or equal to actual PF available"
  exit 1
fi

# This function gets a command and invoke it repeatedly
# until the command return code is zero
function retry {
  local -r tries=$1
  local -r wait_time=$2
  local -r action=$3
  local -r wait_message=$4
  local -r waiting_action=$5

  eval $action
  local return_code=$?
  for i in $(seq $tries); do
    if [[ $return_code -ne 0 ]] ; then
      echo "[$i/$tries] $wait_message"
      eval $waiting_action
      sleep $wait_time
      eval $action
      return_code=$?
    else
      return 0
    fi
  done

  return 1
}

function wait_for_daemonSet {
  local name=$1
  local namespace=$2
  local required_replicas=$3

  if [[ $namespace != "" ]];then
    namespace="-n $namespace"
  fi

  if (( required_replicas < 0 )); then
      echo "DaemonSet $name ready replicas number is not valid: $required_replicas"
      return 1
  fi

  local -r tries=30
  local -r wait_time=10
  wait_message="Waiting for DaemonSet $name to have $required_replicas ready replicas"
  error_message="DaemonSet $name did not have $required_replicas ready replicas"
  action="_kubectl get daemonset $namespace $name -o jsonpath='{.status.numberReady}' | grep -w $required_replicas"

  if ! retry "$tries" "$wait_time" "$action" "$wait_message";then
    echo $error_message
    return 1
  fi

  return  0
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

  if ! retry "$tries" "$wait_time" "$action" "$wait_message";then
    echo $error_message
    return  1
  fi

  return 0
}

function _check_all_pods_ready() {
  all_pods_ready_condition=$(_kubectl get pods -A --no-headers -o custom-columns=':.status.conditions[?(@.type == "Ready")].status')
  if [ "$?" -eq 0 ]; then
    pods_not_ready_count=$(grep -cw False <<< "$all_pods_ready_condition")
    if [ "$pods_not_ready_count" -eq 0 ]; then
      return 0
    fi
  fi

  return 1
}

# not using kubectl wait since with the sriov operator the pods get restarted a couple of times and this is
# more reliable
function wait_pods_ready {
  local -r tries=30
  local -r wait_time=10

  local -r wait_message="Waiting for all pods to become ready.."
  local -r error_message="Not all pods were ready after $(($tries*$wait_time)) seconds"

  local -r get_pods='_kubectl get pods --all-namespaces'
  local -r action="_check_all_pods_ready"

  set +x
  trap "set -x" RETURN

  if ! retry "$tries" "$wait_time" "$action" "$wait_message" "$get_pods"; then
    echo $error_message
    return 1
  fi

  echo "all pods are ready"
  return 0
}

function wait_allocatable_resource {
  local -r node=$1
  local resource_name=$2
  local -r expected_value=$3

  local -r tries=48
  local -r wait_time=10

  local -r wait_message="wait for $node node to have allocatable resource: $resource_name: $expected_value"
  local -r error_message="node $node doesnt have allocatable resource $resource_name:$expected_value"

  # it is necessary to add '\' before '.' in the resource name.
  resource_name=$(echo $resource_name | sed s/\\./\\\\\./g)
  local -r action='_kubectl get node $node -ocustom-columns=:.status.allocatable.$resource_name --no-headers | grep -w $expected_value'

  if ! retry $tries $wait_time "$action" "$wait_message"; then
    echo $error_message

    echo "LOGS network-resources-injector"
    POD=$(_kubectl get pods -n sriov-network-operator | grep network-resources-injector | awk '{print $1}')
    _kubectl logs -n sriov-network-operator $POD

    echo "LOGS operator-webhook"
    POD=$(_kubectl get pods -n sriov-network-operator | grep operator-webhook | awk '{print $1}')
    _kubectl logs -n sriov-network-operator $POD

    echo "LOGS sriov-cni"
    POD=$(_kubectl get pods -n sriov-network-operator | grep sriov-cni | awk '{print $1}')
    _kubectl logs -n sriov-network-operator $POD

    echo "LOGS sriov-device-plugin"
    POD=$(_kubectl get pods -n sriov-network-operator | grep sriov-device-plugin | awk '{print $1}')
    _kubectl logs -n sriov-network-operator $POD

    echo "LOGS sriov-network-config-daemon"
    POD=$(_kubectl get pods -n sriov-network-operator | grep sriov-network-config-daemon | awk '{print $1}')
    _kubectl logs -n sriov-network-operator $POD

    echo "LOGS sriov-network-operator"
    POD=$(_kubectl get pods -n sriov-network-operator | grep sriov-network-operator | awk '{print $1}')
    _kubectl logs -n sriov-network-operator $POD

    return 1
  fi

  return 0
}

function deploy_multus {
  echo 'Deploying Multus'
  _kubectl create -f $MANIFESTS_DIR/multus.yaml

  echo 'Waiting for Multus deployment to become ready'
  daemonset_name=$(cat $MANIFESTS_DIR/multus.yaml | grep -i daemonset -A 3 | grep -Po '(?<=name:) \S*amd64$')
  daemonset_namespace=$(cat $MANIFESTS_DIR/multus.yaml | grep -i daemonset -A 3 | grep -Po '(?<=namespace:) \S*$' | head -1)
  required_replicas=$(_kubectl get daemonset $daemonset_name -n $daemonset_namespace -o jsonpath='{.status.desiredNumberScheduled}')
  wait_for_daemonSet $daemonset_name $daemonset_namespace $required_replicas

  return 0
}

function deploy_sriov_operator {
  echo 'Downloading the SR-IOV operator'
  operator_path=${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/sriov-network-operator-${OPERATOR_GIT_HASH}
  if [ ! -d $operator_path ]; then
    curl -LSs https://github.com/openshift/sriov-network-operator/archive/${OPERATOR_GIT_HASH}/sriov-network-operator.tar.gz | tar xz -C ${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/
  fi

  export SRIOV_NETWORK_OPERATOR_IMAGE=quay.io/openshift/origin-sriov-network-operator:${RELEASE_VERSION}
  for ifs in "${sriov_pfs_totalvfs[@]}"; do
    ifs="${ifs%%/sriov_totalvfs}"
    export IGNORE_PATH=$ifs

    docker pull $SRIOV_NETWORK_OPERATOR_IMAGE
    TAG=$(docker create $SRIOV_NETWORK_OPERATOR_IMAGE)
    docker cp $TAG:/bindata/manifests/daemon/daemonset.yaml daemonset.yaml

    daemonset_diff="$MANIFESTS_DIR/daemonset.diff.template"
    envsubst < $daemonset_diff > daemonset.diff
    patch daemonset.yaml daemonset.diff
    docker cp daemonset.yaml $TAG:/bindata/manifests/daemon/daemonset.yaml
    docker commit $TAG localhost:$HOST_PORT/kubevirt/patched_sriov_operator:${RELEASE_VERSION}
    docker push localhost:$HOST_PORT/kubevirt/patched_sriov_operator:${RELEASE_VERSION}

    docker rm $TAG
    export SRIOV_NETWORK_OPERATOR_IMAGE=registry:5000/kubevirt/patched_sriov_operator:${RELEASE_VERSION}

    break # because there can be just one for now
  done

  echo 'Installing the SR-IOV operator'
  pushd $operator_path
    export SKIP_VAR_SET=1
    export CGO_ENABLED=0
    export SRIOV_NETWORK_CONFIG_DAEMON_IMAGE=quay.io/openshift/origin-sriov-network-config-daemon:${RELEASE_VERSION}
    export SRIOV_NETWORK_WEBHOOK_IMAGE=quay.io/openshift/origin-sriov-network-webhook:${RELEASE_VERSION}
    export NETWORK_RESOURCES_INJECTOR_IMAGE=quay.io/openshift/origin-sriov-dp-admission-controller:${RELEASE_VERSION}
    export SRIOV_CNI_IMAGE=quay.io/openshift/origin-sriov-cni:${RELEASE_VERSION}
    export SRIOV_DEVICE_PLUGIN_IMAGE=quay.io/openshift/origin-sriov-network-device-plugin:${RELEASE_VERSION}
    export SRIOV_INFINIBAND_CNI_IMAGE=quay.io/openshift/origin-sriov-infiniband-cni:${RELEASE_VERSION}

    # use deploy-setup in order to avoid eliminating webhook creation by deploy-setup-k8s
    export NAMESPACE=sriov-network-operator
    export ENABLE_ADMISSION_CONTROLLER="true"
    export CNI_BIN_PATH=/opt/cni/bin
    export OPERATOR_EXEC=${KUBECTL}
    # on prow nodes the default shell is dash and some commands are not working
    make deploy-setup SHELL=/bin/bash
  popd

  echo 'Generating webhook certificates for the SR-IOV operator webhooks'
  pushd "${CERTCREATOR_PATH}"
    go run . -namespace sriov-network-operator -secret operator-webhook-service -hook operator-webhook -kubeconfig $KUBECONFIG_PATH
    go run . -namespace sriov-network-operator -secret network-resources-injector-secret -hook network-resources-injector -kubeconfig $KUBECONFIG_PATH
  popd

  echo 'Setting caBundle for SR-IOV webhooks'
  wait_k8s_object "validatingwebhookconfiguration" "sriov-operator-webhook-config"
  _kubectl patch validatingwebhookconfiguration sriov-operator-webhook-config --patch '{"webhooks":[{"name":"operator-webhook.sriovnetwork.openshift.io", "clientConfig": { "caBundle": "'"$(cat $CERTCREATOR_PATH/operator-webhook.cert)"'" }}]}'

  wait_k8s_object "mutatingwebhookconfiguration" "sriov-operator-webhook-config"
  _kubectl patch mutatingwebhookconfiguration sriov-operator-webhook-config --patch '{"webhooks":[{"name":"operator-webhook.sriovnetwork.openshift.io", "clientConfig": { "caBundle": "'"$(cat $CERTCREATOR_PATH/operator-webhook.cert)"'" }}]}'

  wait_k8s_object "mutatingwebhookconfiguration" "network-resources-injector-config"
  _kubectl patch mutatingwebhookconfiguration network-resources-injector-config --patch '{"webhooks":[{"name":"network-resources-injector-config.k8s.io", "clientConfig": { "caBundle": "'"$(cat $CERTCREATOR_PATH/network-resources-injector.cert)"'" }}]}'

  return 0
}

function apply_sriov_node_policy {
  local -r policy_file=$1

  # Substitute $NODE_PF and $NODE_PF_NUM_VFS and create SriovNetworkNodePolicy CR
  local -r policy=$(envsubst < $policy_file)
  echo "Applying SriovNetworkNodeConfigPolicy:"
  echo "$policy"

  # until https://github.com/k8snetworkplumbingwg/sriov-network-operator/issues/3 is fixed we need to inject CaBundle and retry policy creation
  tries=0
  until _kubectl create -f - <<< "$policy"; do
    if [ $tries -eq 10 ]; then
      echo "could not create policy"
      return 1
    fi
    _kubectl patch validatingwebhookconfiguration sriov-operator-webhook-config --patch '{"webhooks":[{"name":"operator-webhook.sriovnetwork.openshift.io", "clientConfig": { "caBundle": "'"$(cat $CERTCREATOR_PATH/operator-webhook.cert)"'" }}]}'
    _kubectl patch mutatingwebhookconfiguration sriov-operator-webhook-config --patch '{"webhooks":[{"name":"operator-webhook.sriovnetwork.openshift.io", "clientConfig": { "caBundle": "'"$(cat $CERTCREATOR_PATH/operator-webhook.cert)"'" }}]}'
    _kubectl patch mutatingwebhookconfiguration network-resources-injector-config --patch '{"webhooks":[{"name":"network-resources-injector-config.k8s.io", "clientConfig": { "caBundle": "'"$(cat $CERTCREATOR_PATH/network-resources-injector.cert)"'" }}]}'
    tries=$((tries+1))
  done

  return 0
}

function setns_sriov_ifs {
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

  PF_COUNTER=0
  # Scan available sriov PFs
  sriov_pfs=( /sys/class/net/*/device/sriov_numvfs )
  if [ $sriov_pfs == "/sys/class/net/*/device/sriov_numvfs" ]; then
    echo "FATAL: No sriov PFs found"
    exit 1
  fi

  sriov_pfs_totalvfs=( $(find /sys/devices -name sriov_totalvfs 2>/dev/null) )
  for ifs in "${sriov_pfs_totalvfs[@]}"; do
    echo $ifs
  done

  for ifs in "${sriov_pfs[@]}"; do
    ifs_name="${ifs%%/device/*}"
    ifs_name="${ifs_name##*/}"

    if [ $(echo "${PF_BLACKLIST[@]}" | grep "${ifs_name}") ]; then
      continue
    fi

    export PF_ADDRESS=$(cat /sys/class/net/$ifs_name/device/uevent | grep PCI_SLOT_NAME | cut -d= -f2)
    export tmp_pf_num_vfs=$(cat /sys/class/net/"$ifs_name"/device/sriov_totalvfs)

    # In case two clusters started at the same time, they might race on the same PF.
    # The first will manage to assign the PF to its container, and the 2nd will just skip it
    # and try the rest of the PFs available.
    if ip link set "$ifs_name" netns "$SRIOV_NODE"; then
      if timeout 5s bash -c "until docker exec $SRIOV_NODE ip address | grep -w $ifs_name; do sleep 1; done"; then
        # We set the variable below only in the first iteration as we need only one PF
        # to inject into the Network Configuration manifest. We need to move all pfs to
        # the node's namespace and for that reason we do not interrupt the loop.
        if [ -z "$NODE_PF" ]; then
          # These values are used to populate the network definition policy yaml.
          # We just use the first suitable pf
          # We need the num of vfs because if we don't set this value equals to the total, in case of mellanox
          # the sriov operator will trigger a node reboot to update the firmware
          export NODE_PF="$ifs_name"
          export NODE_PF_NUM_VFS=$tmp_pf_num_vfs
        fi

        for index in "${!sriov_pfs_totalvfs[@]}"; do
          [ $(grep $PF_ADDRESS"/sriov_totalvfs" <<< ${sriov_pfs_totalvfs[index]}) ] && unset -v 'sriov_pfs_totalvfs[$index]'
        done

        PF_COUNTER=$((PF_COUNTER+1))
        if [[ $PF_COUNTER -eq $NUM_PF_REQUIRED ]]; then
          echo "Allocated requested number of PFs"
          break
        fi
      fi
    fi
  done

  if [[ $PF_COUNTER -lt $NUM_PF_REQUIRED ]]; then
    echo "FATAL: Could not allocate enough PFs, please check PF_BLACKLIST, NUM_PF_REQUIRED, and how many PF actually available or used already"
    exit 1
  fi
}

setns_sriov_ifs

lspci -i $NODE_PF -vvvv

SRIOV_NODE_CMD="docker exec -it -d ${SRIOV_NODE}"
${SRIOV_NODE_CMD} mount -o remount,rw /sys     # kind remounts it as readonly when it starts, we need it to be writeable
${SRIOV_NODE_CMD} chmod 666 /dev/vfio/vfio
_kubectl label node $SRIOV_NODE sriov=true

TOTAL_PF=$((${#sriov_pfs_totalvfs[@]} + $NUM_PF_REQUIRED))
if [ $TOTAL_PF != 2 ]; then
   echo "Warning currently supporting only 2 PFs total for the POC"
fi

deploy_multus
wait_pods_ready

deploy_sriov_operator
wait_pods_ready

policy="$MANIFESTS_DIR/network_config_policy.yaml"
apply_sriov_node_policy "$policy"

# Verify that sriov node has sriov VFs allocatable resource
resource_name=$(sed -n 's/.*resourceName: *//p' $policy)
wait_allocatable_resource $SRIOV_NODE "openshift.io/$resource_name" $NODE_PF_NUM_VFS
wait_pods_ready

_kubectl get nodes
_kubectl get pods -n $SRIOV_OPERATOR_NAMESPACE
echo
echo "$KUBEVIRT_PROVIDER cluster is ready"
