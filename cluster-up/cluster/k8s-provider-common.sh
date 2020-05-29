#!/usr/bin/env bash

set -e

source ${KUBEVIRTCI_PATH}/cluster/ephemeral-provider-common.sh

function up() {
    ${_cli} run $(_add_common_params)

    # Copy k8s config and kubectl
    ${_cli} scp --prefix $provider_prefix /usr/bin/kubectl - >${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kubectl
    chmod u+x ${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kubectl
    ${_cli} scp --prefix $provider_prefix /etc/kubernetes/admin.conf - >${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kubeconfig

    # Set server and disable tls check
    export KUBECONFIG=${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kubeconfig
    ${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kubectl config set-cluster kubernetes --server=https://$(_main_ip):$(_port k8s)
    ${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kubectl config set-cluster kubernetes --insecure-skip-tls-verify=true

    # Make sure that local config is correct
    prepare_config


    kubectl="${_cli} --prefix $provider_prefix ssh node01 -- sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf"

    # For multinode cluster Label all the non master nodes as workers,
    # for one node cluster label master with 'master,worker' roles
    if [ "$KUBEVIRT_NUM_NODES" -gt 1 ]; then
        label="!node-role.kubernetes.io/master"
    else
        label="node-role.kubernetes.io/master"
    fi
    $kubectl label node -l $label node-role.kubernetes.io/worker=''

    # Activate cluster-network-addons-operator if flag is passed
    if [ "$KUBEVIRT_WITH_CNAO" == "true" ]; then

        $kubectl create -f /opt/cnao/namespace.yaml
        $kubectl create -f /opt/cnao/network-addons-config.crd.yaml
        $kubectl create -f /opt/cnao/operator.yaml
        $kubectl wait deployment -n cluster-network-addons cluster-network-addons-operator --for condition=Available --timeout=200s

        $kubectl create -f /opt/cnao/network-addons-config-example.cr.yaml
        $kubectl wait networkaddonsconfig cluster --for condition=Available --timeout=200s
    fi

    if [ "$KUBEVIRT_OVS_DPDK" == "true" ]; then
        OVSDPDK_MANIFESTS_DIR="${KUBEVIRTCI_PATH}/cluster/$KUBEVIRT_PROVIDER/manifests/ovsdpdk"
        if [ -f $OVSDPDK_MANIFESTS_DIR/allinone.yaml ]; then
            echo "Create OvS-DPDK operator.."
            _kubectl label node node02 network.operator.openshift.io/external-openvswitch=
            _kubectl create -f $OVSDPDK_MANIFESTS_DIR/allinone.yaml
            _kubectl create -f $OVSDPDK_MANIFESTS_DIR/ovsdpdkconfig.yaml
        fi
    fi
}
