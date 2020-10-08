#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -exuo pipefail

(cd ../gocli && make container)

CI=${CI:-"false"}
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
provision_dir="$1"

function cleanup {
  cd "$DIR" && cd ../..
  make cluster-down
}

export KUBEVIRTCI_GOCLI_CONTAINER=kubevirtci/gocli:devel
# check cluster-up
(
  ksh="./cluster-up/kubectl.sh"
  cd "$DIR" && cd ../..
  export KUBEVIRTCI_PROVISION_CHECK=1
  export KUBEVIRT_PROVIDER="k8s-${provision_dir}"
  export KUBEVIRT_NUM_NODES=2
  export KUBEVIRT_NUM_SECONDARY_NICS=2
  trap cleanup EXIT ERR SIGINT SIGTERM SIGQUIT
  bash -x ./cluster-up/up.sh
  ${ksh} wait --for=condition=Ready pod --timeout=200s --all
  ${ksh} wait --for=condition=Ready pod --timeout=200s -n kube-system --all
  ${ksh} get nodes
  ${ksh} get pods -A

  # Run conformance test only at CI and if the provider has them activated
  conformance_config=$DIR/${provision_dir}/conformance.json
  if [ "${CI}" == "true" -a -f $conformance_config ]; then
    hack/conformance.sh $conformance_config
  fi
)
