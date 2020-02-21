#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -exuo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
provision_dir="$1"

function cleanup {
  cd "$DIR" && cd ../..
  make cluster-down
}

# check cluster-up
(
  ksh="./cluster-up/kubectl.sh"
  cd "$DIR" && cd ../..
  export KUBEVIRTCI_PROVISION_CHECK=1
  export KUBEVIRT_PROVIDER="k8s-${provision_dir}"
  export KUBEVIRT_NUM_NODES=2
  trap cleanup EXIT ERR SIGINT SIGTERM SIGQUIT
  bash -x ./cluster-up/up.sh
  ${ksh} wait --for=condition=Ready pod --all
  ${ksh} wait --for=condition=Ready pod -n kube-system --all
  ${ksh} get nodes
  ${ksh} get pods -A
)
