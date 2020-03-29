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

    kubectl=${KUBEVIRTCI_PATH}/kubectl.sh

    # Label all the non master nodes as workers, we have to do here since, after k8s 1.16 is not possible to do
    # at kubelet [1]
    # [1] https://github.com/kubernetes-sigs/cluster-api/blob/master/docs/book/src/user/troubleshooting.md
    $kubectl label node -l '!node-role.kubernetes.io/master' node-role.kubernetes.io/worker=''
    # Activate cluster-network-addons-operator if flag is passed
    if [ "$KUBEVIRT_WITH_CNAO" == "true" ]; then

        $kubectl create -f /opt/cnao/namespace.yaml
        $kubectl create -f /opt/cnao/network-addons-config.crd.yaml
        $kubectl create -f /opt/cnao/operator.yaml
        $kubectl wait deployment -n cluster-network-addons cluster-network-addons-operator --for condition=Available --timeout=200s

        $kubectl create -f /opt/cnao/network-addons-config-example.cr.yaml
        $kubectl wait networkaddonsconfig cluster --for condition=Available --timeout=200s
    fi
}
