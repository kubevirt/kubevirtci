#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -exuo pipefail

make -C ../gocli container

CI=${CI:-"false"}
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SLIM=${SLIM:-false}
RUN_KUBEVIRT_CONFORMANCE=${RUN_KUBEVIRT_CONFORMANCE:-"true"}

provision_dir="$1"
provider="${provision_dir}"

function cleanup() {
    cd "$DIR" && cd ../..
    make cluster-down
}

export KUBEVIRTCI_GOCLI_CONTAINER=quay.io/kubevirtci/gocli:latest
# check cluster-up
(
    ksh="./cluster-up/kubectl.sh"
    ssh="./cluster-up/ssh.sh"
    cd "$DIR" && cd ../..
    export KUBEVIRTCI_PROVISION_CHECK=1
    export KUBEVIRT_PROVIDER="k8s-${provider}"
    export KUBEVIRT_NUM_NODES=2
    export KUBEVIRT_MEMORY_SIZE=5520M
    export KUBEVIRT_NUM_SECONDARY_NICS=2
    export KUBEVIRT_WITH_CNAO=true
    export KUBEVIRT_WITH_MULTUS_V3=true
    export KUBEVIRT_DEPLOY_ISTIO=true
    export KUBEVIRT_DEPLOY_PROMETHEUS=true
    export KUBEVIRT_DEPLOY_PROMETHEUS_ALERTMANAGER=true
    export KUBEVIRT_DEPLOY_GRAFANA=true
    export KUBEVIRT_DEPLOY_CDI=true

    trap cleanup EXIT ERR SIGINT SIGTERM SIGQUIT
    bash -x ./cluster-up/up.sh
    timeout 210s bash -c "until ${ksh} wait --for=condition=Ready pod --timeout=30s --all -l app!=whereabouts; do sleep 1; done"
    timeout 210s bash -c "until ${ksh} wait --for=condition=Ready pod --timeout=30s -n kube-system --all -l app!=whereabouts; do sleep 1; done"
    ${ksh} get nodes
    ${ksh} get pods -A -owide

    # Run some checks for KUBEVIRT_NUM_NODES
    # and KUBEVIRT_NUM_SECONDARY_NICS
    ${ksh} get node node01
    ${ksh} get node node02
    ${ssh} node01 -- ip l show eth1
    ${ssh} node01 -- ip l show eth2
    ${ssh} node02 -- ip l show eth1
    ${ssh} node02 -- ip l show eth2

    # Verify Multus v3 image is used
    ${ksh} get ds -n kube-system kube-multus-ds -o yaml | grep multus-cni:v3

    # Sanity check that Multus able to connect secondary networks
    ${ksh} create -f "$DIR/test-multi-net.yaml"
    ${ksh} wait pod test-multi-net --for condition=ready=true
    ${ksh} delete -f "$DIR/test-multi-net.yaml"

    pre_pull_image_file="$DIR/${provision_dir}/extra-pre-pull-images"
    if [ -f "${pre_pull_image_file}" ]; then
        bash -x "$DIR/deploy-manifests.sh" "${provision_dir}"
        bash -x "$DIR/validate-pod-pull-policies.sh"
        if [[ ${SLIM} == false ]]; then
            bash -x "$DIR/check-pod-images.sh" "${provision_dir}"
        fi
    fi

    # Run conformance test only at CI and if the provider has them activated
    conformance_config=$DIR/${provision_dir}/conformance.json

    if [ "${CI}" == "true" ] && [ -f $conformance_config ]; then
        if [ "$RUN_KUBEVIRT_CONFORMANCE" == "true" ]; then
            LATEST=$(curl -L "https://storage.googleapis.com/kubevirt-prow/devel/nightly/release/kubevirt/kubevirt/latest")
            ${ksh} apply -f "https://storage.googleapis.com/kubevirt-prow/devel/nightly/release/kubevirt/kubevirt/${LATEST}/kubevirt-operator.yaml"
            ${ksh} apply -f "https://storage.googleapis.com/kubevirt-prow/devel/nightly/release/kubevirt/kubevirt/${LATEST}/kubevirt-cr.yaml"

            ${ksh} wait -n kubevirt kv kubevirt --for condition=Available --timeout 15m

            if [ "${KUBEVIRT_PSA:-"false"}" == "true" ]; then
                # Enable Kubevirt profile
                ${ksh} patch -n kubevirt kv kubevirt --type='merge' --patch '{"spec": {"configuration": {"developerConfiguration": {"featureGates": ["KubevirtSeccompProfile"]} } }}'
                ${ksh} patch -n kubevirt kv kubevirt --type='merge' --patch '{"spec": {"configuration": {"seccompConfiguration": {"virtualMachineInstanceProfile": {"customProfile": {"localhostProfile" : "kubevirt/kubevirt.json"} } } } } }'
            fi

            export SONOBUOY_EXTRA_ARGS="--plugin https://storage.googleapis.com/kubevirt-prow/devel/nightly/release/kubevirt/kubevirt/${LATEST}/conformance.yaml"
            hack/conformance.sh $conformance_config
        fi

        export SONOBUOY_EXTRA_ARGS="--plugin systemd-logs --plugin e2e"
        hack/conformance.sh $conformance_config

        echo "Sanity check cluster-up of single stack cluster"
        make cluster-down
        export KUBEVIRT_SINGLE_STACK=true
        make cluster-up
    fi
)
