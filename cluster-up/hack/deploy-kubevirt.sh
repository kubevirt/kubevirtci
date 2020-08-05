#!/bin/bash
set -exo pipefail

# Deploys KubeVirt latest stable release
#
# Example:
#   export KUBECONFIG=< path to cluster token >
#   ./deploy-kubevirt
# In order to deploy specific version:
#   KUBEVIRT_VERSION="v0.tries.0" ./deploy-kubevirt.sh


readonly kubevirt_releases_url="https://api.github.com/repos/kubevirt/kubevirt/releases"
if [ ! -f "${KUBECONFIG}" ];then
  echo ".kubeconfig file not found at: '$KUBECONFIG'" && exit 1
fi

if [ -z "${KUBEVIRT_VERSION}" ];then
  # Get latest stable KubeVirt version
  export KUBEVIRT_VERSION=$(curl -s $kubevirt_releases_url | grep tag_name | grep -v -- - | sort -V | tail -1 | awk -F':' '{print $2}' | sed 's/,//' | xargs)
fi

function _kubectl() {
  ./cluster/kubectl.sh "$@"
}

kubevirt_operator_manifest="https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-operator.yaml"
kubevirt_cr_manifest="https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-cr.yaml"

_kubectl apply -f $kubevirt_operator_manifest

# Ensure the KubeVirt CRD is created
count=0
tries=30
wait_time=1
until _kubectl get crd kubevirts.kubevirt.io; do
    ((count++)) && ((count == tries)) && echo "KubeVirt CRD not found" && exit 1
    echo "waiting for KubeVirt CRD"
    sleep $wait_time
done

_kubectl apply -f $kubevirt_cr_manifest

# Ensure the KubeVirt CR is created
count=0
tries=30
wait_time=1
until _kubectl -n kubevirt get kv kubevirt; do
    ((count++)) && ((count == tries)) && echo "KubeVirt CR not found" && exit 1
    echo "waiting for KubeVirt CR"
    sleep $wait_time
done

echo "Waiting for all KubeVirt pods to be ready"
_kubectl wait -n kubevirt kv kubevirt --for condition=Available --timeout 180s || (echo "KubeVirt not ready in time" && exit 1)

echo "KubeVirt deployment finished successful"
_kubectl get pods -n kubevirt

echo "basename -- $0 done"
