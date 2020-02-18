#!/usr/bin/env bash

function _kubectl() {
    kubectl "$@"
}

function prepare_config() {
    BASE_PATH=${KUBEVIRTCI_CONFIG_PATH:-$PWD}

    # required for running tests within openshift ci, otherwise tests don't find the configuration
    ln -f -s "${KUBECONFIG:-$HOME/.kube/config}" "${BASE_PATH}/$KUBEVIRT_PROVIDER/.kubeconfig"

    cat > "${BASE_PATH}/$KUBEVIRT_PROVIDER/config-provider-$KUBEVIRT_PROVIDER.sh" <<EOF
kubeconfig=${BASE_PATH}/$KUBEVIRT_PROVIDER/.kubeconfig
docker_tag=\${DOCKER_TAG}
docker_prefix=\${DOCKER_PREFIX}
manifest_docker_prefix=\${DOCKER_PREFIX}
image_pull_policy=\${IMAGE_PULL_POLICY:-Always}
EOF
}

# The external cluster is assumed to be up.  Do a simple check
function up() {
    prepare_config
    if ! _kubectl version >/dev/null; then
        echo -e "\n*** Unable to reach external cluster.  Please check configuration ***"
        echo -e "*** Type \"kubectl config view\" for current settings               ***\n"
        exit 1
    fi
    echo "Cluster is up"
}

function down() {
    echo "Not supported by this provider"
}

