unset docker_prefix master_ip network_provider kubeconfig manifest_docker_prefix

KUBEVIRT_PROVIDER=${KUBEVIRT_PROVIDER:-${PROVIDER}}

source ${KUBEVIRTCI_PATH}hack/config-default.sh

# Allow different providers to override default config values
test -f "${KUBEVIRTCI_PATH}hack/config-provider-${KUBEVIRT_PROVIDER}.sh" && source ${KUBEVIRTCI_PATH}hack/config-provider-${KUBEVIRT_PROVIDER}.sh

# Let devs override any default variables, to avoid needing
# to change the version controlled config-default.sh file
test -f "${KUBEVIRTCI_PATH}hack/config-local.sh" && source ${KUBEVIRTCI_PATH}hack/config-local.sh

export docker_prefix master_ip network_provider kubeconfig manifest_docker_prefix
