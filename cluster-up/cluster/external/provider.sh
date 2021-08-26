#!/usr/bin/env bash

KUBECTL="${KUBEVIRTCI_CONFIG_PATH}/$KUBEVIRT_PROVIDER/.kubectl --kubeconfig=${KUBECONFIG}"

function get_kubectl() {
    PLATFORM=$(uname -m)
    case ${PLATFORM} in
    x86_64* | i?86_64* | amd64*)
        ARCH="amd64"
        ;;
    aarch64* | arm64*)
        ARCH="arm64"
        ;;
    ppc64le)
        ARCH="ppc64le"
        ;;
    *)
        echo "invalid Arch, only support x86_64, aarch64 and ppc64le"
        exit 1
        ;;
    esac

    if [ -z "${KUBEVERSION}" ]; then
        echo "KUBEVERSION is not set!"
        exit 1
    fi

    curl --fail -L "https://dl.k8s.io/release/${KUBEVERSION}/bin/linux/${ARCH}/kubectl" -o ${BASE_PATH}/$KUBEVIRT_PROVIDER/.kubectl
    if [ $? -ne 0 ];then
        echo "invalid kubectl binary download address"
        echo "https://dl.k8s.io/release/${KUBEVERSION}/bin/linux/${ARCH}/kubectl"
        exit 1
    fi
    chmod +x ${BASE_PATH}/$KUBEVIRT_PROVIDER/.kubectl
}

function _kubectl() {
    ${KUBECTL} "$@"
}

function prepare_config() {
    BASE_PATH=${KUBEVIRTCI_CONFIG_PATH:-$PWD}
    get_kubectl

    if [ -z "${KUBECONFIG}" ]; then
        echo "KUBECONFIG is not set!"
        exit 1
    fi

    PROVIDER_CONFIG_FILE_PATH="${BASE_PATH}/$KUBEVIRT_PROVIDER/config-provider-$KUBEVIRT_PROVIDER.sh"

    cat > "$PROVIDER_CONFIG_FILE_PATH" <<EOF
kubeconfig=\${KUBECONFIG}
kubectl=${BASE_PATH}/$KUBEVIRT_PROVIDER/.kubectl
docker_tag=\${DOCKER_TAG}
docker_prefix=\${DOCKER_PREFIX}
manifest_docker_prefix=\${DOCKER_PREFIX}
image_pull_policy=\${IMAGE_PULL_POLICY:-Always}
EOF

    if which oc; then
        echo "oc=$(which oc)" >> "$PROVIDER_CONFIG_FILE_PATH"
    fi

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

