#!/usr/bin/env bash

set -e

if [ "${KUBEVIRTCI_RUNTIME}" = "podman" ]; then
    _cli="pack8s"
else
    _cli_container="${KUBEVIRTCI_GOCLI_CONTAINER:-kubevirtci/gocli@sha256:3cebe61e23e75bb6342530237ebc40ead4db19f632afeedc812852151498b620}"
    _cli="docker run --privileged --net=host --rm ${USE_TTY} -v /var/run/docker.sock:/var/run/docker.sock"
    # gocli will try to mount /lib/modules to make it accessible to dnsmasq in
    # in case it exists
    if [ -d /lib/modules ]; then
        _cli="${_cli} -v /lib/modules/:/lib/modules/"
    fi
    _cli="${_cli} ${_cli_container}"
fi

function _main_ip() {
    echo 127.0.0.1
}

function _port() {
    ${_cli} ports --prefix $provider_prefix "$@"
}

function prepare_config() {
    BASE_PATH=${KUBEVIRTCI_CONFIG_PATH:-$PWD}
    cat >$BASE_PATH/$KUBEVIRT_PROVIDER/config-provider-$KUBEVIRT_PROVIDER.sh <<EOF
master_ip=$(_main_ip)
kubeconfig=${BASE_PATH}/$KUBEVIRT_PROVIDER/.kubeconfig
kubectl=${BASE_PATH}/$KUBEVIRT_PROVIDER/.kubectl
gocli=${BASE_PATH}/../cluster-up/cli.sh
docker_prefix=\${DOCKER_PREFIX:-localhost:$(_port registry)/kubevirt}
manifest_docker_prefix=\${DOCKER_PREFIX:-registry:5000/kubevirt}
EOF
}

function _registry_volume() {
    echo ${job_prefix}_registry
}

function _add_common_params() {
    local params="--nodes ${KUBEVIRT_NUM_NODES} --memory ${KUBEVIRT_MEMORY_SIZE} --cpu 6 --secondary-nics ${KUBEVIRT_NUM_SECONDARY_NICS} --random-ports --background --prefix $provider_prefix --registry-volume $(_registry_volume) ${KUBEVIRT_PROVIDER} ${KUBEVIRT_PROVIDER_EXTRA_ARGS}"
    if [[ $TARGET =~ windows.* ]] && [ -n "$WINDOWS_NFS_DIR" ]; then
        params=" --nfs-data $WINDOWS_NFS_DIR $params"
    elif [[ $TARGET =~ os-.* ]] && [ -n "$RHEL_NFS_DIR" ]; then
        params=" --nfs-data $RHEL_NFS_DIR $params"
    fi
    if [ -n "${KUBEVIRTCI_PROVISION_CHECK}" ]; then
        params=" --container-registry= --container-suffix=:latest $params"
    fi
    echo $params
}

function _kubectl() {
    export KUBECONFIG=${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kubeconfig
    ${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kubectl "$@"
}

function down() {
    ${_cli} rm --prefix $provider_prefix
}
