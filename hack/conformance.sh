set -e

ARTIFACTS=${ARTIFACTS:-${PWD}}

if [[ -z "$KUBEVIRT_PROVIDER" ]]; then
    echo "KUBEVIRT_PROVIDER is not set" 1>&2
    exit 1
fi

curl -O -L https://github.com/vmware-tanzu/sonobuoy/releases/download/v0.18.2/sonobuoy_0.18.2_linux_amd64.tar.gz
tar xf sonobuoy_0.18.2_linux_amd64.tar.gz

export KUBECONFIG=_ci-configs/${KUBEVIRT_PROVIDER}/.kubeconfig
trap "./sonobuoy status --json; ./sonobuoy logs > ${ARTIFACTS}/sonobuoy.log" EXIT
./sonobuoy run --wait
