#!/usr/bin/env bash

function _kubectl() {
    kubectl "$@"
}

function prepare_config() {
    export kubeconfig=${KUBECONFIG}
    export docker_tag=${DOCKER_TAG}
    export docker_prefix=${DOCKER_PREFIX}
    export manifest_docker_prefix=${DOCKER_PREFIX}
    export image_pull_policy=\${IMAGE_PULL_POLICY:-Always}
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

