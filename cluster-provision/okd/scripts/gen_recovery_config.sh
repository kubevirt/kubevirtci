#!/bin/bash -xe

function start_recovery_api_server() {
    RELEASE_IMAGE=docker.io/kubevirtci/ocp-release:4.1.24
    KAO_IMAGE=$(oc adm release info --registry-config='/var/lib/kubelet/config.json' "${RELEASE_IMAGE}" --image-for=cluster-kube-apiserver-operator)
    podman pull --authfile=/var/lib/kubelet/config.json "${KAO_IMAGE}"
    podman run -it --network=host -v /etc/kubernetes/:/etc/kubernetes/:Z --entrypoint=/usr/bin/cluster-kube-apiserver-operator "${KAO_IMAGE}" recovery-apiserver create
    export KUBECONFIG=/etc/kubernetes/static-pod-resources/recovery-kube-apiserver-pod/admin.kubeconfig
    until oc get namespace kube-system 2>/dev/null 1>&2; do
        echo 'Waiting for recovery apiserver to come up.'
        sleep 1
    done
}

function regenerate_certificates() {
    podman run -it --network=host -v /etc/kubernetes/:/etc/kubernetes/:Z --entrypoint=/usr/bin/cluster-kube-apiserver-operator "${KAO_IMAGE}" regenerate-certificates
    oc patch kubeapiserver cluster -p='{"spec": {"forceRedeploymentReason": "recovery-'"$( date --rfc-3339=ns )"'"}}' --type=merge
    oc patch kubecontrollermanager cluster -p='{"spec": {"forceRedeploymentReason": "recovery-'"$( date --rfc-3339=ns )"'"}}' --type=merge
    oc patch kubescheduler cluster -p='{"spec": {"forceRedeploymentReason": "recovery-'"$( date --rfc-3339=ns )"'"}}' --type=merge
}

function create_bootstrap_config() {
    /usr/local/bin/recover-kubeconfig.sh > /tmp/recovery-kubeconfig
    cp /tmp/recovery-kubeconfig /etc/kubernetes/kubeconfig
}

function generate_kubeapi_ca_cert() {
    oc get configmap kube-apiserver-to-kubelet-client-ca -n openshift-kube-apiserver-operator --template='{{ index .data "ca-bundle.crt" }}' > /tmp/kubernetes-ca.crt
    cp /tmp/kubernetes-ca.crt /etc/kubernetes/ca.crt
}

function main() {
    start_recovery_api_server
    generate_recovery_kubeconfig
    create_bootstrap_config
    generate_kubeapi_ca_cert
}