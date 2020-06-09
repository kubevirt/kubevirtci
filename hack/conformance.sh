set -xe

ARTIFACTS=${ARTIFACTS:-${PWD}}
sonobuoy_version=0.18.2

if [[ -z "$KUBEVIRT_PROVIDER" ]]; then
    echo "KUBEVIRT_PROVIDER is not set" 1>&2
    exit 1
fi

export KUBECONFIG=$(cluster-up/kubeconfig.sh)

teardown() {
    ./sonobuoy status --json
    ./sonobuoy logs > ${ARTIFACTS}/sonobuoy.log
    results_tarball=$(./sonobuoy retrieve)
    tar -xvzf $results_tarball plugins/e2e/results/
    cp -f $(find plugins/e2e/results/* -name "*.xml") ${ARTIFACTS}/
}

curl -L https://github.com/vmware-tanzu/sonobuoy/releases/download/v${sonobuoy_version}/sonobuoy_${sonobuoy_version}_linux_amd64.tar.gz | tar -xz

trap teardown EXIT
./sonobuoy run --wait
