#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -exuo pipefail

make -C ../gocli container

CI=${CI:-"false"}
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
provision_dir="$1"

function cleanup {
  cd "$DIR" && cd ../..
  make cluster-down
}

export KUBEVIRTCI_GOCLI_CONTAINER=quay.io/kubevirtci/gocli:latest
# check cluster-up
(
  ksh="./cluster-up/kubectl.sh"
  ssh="./cluster-up/ssh.sh"
  cd "$DIR" && cd ../..
  export KUBEVIRTCI_PROVISION_CHECK=1
  export KUBEVIRT_PROVIDER="k8s-${provision_dir}"
  export KUBEVIRT_NUM_NODES=2
  export KUBEVIRT_NUM_SECONDARY_NICS=2
  trap cleanup EXIT ERR SIGINT SIGTERM SIGQUIT
  bash -x ./cluster-up/up.sh
  timeout 210s bash -c "until ${ksh} wait --for=condition=Ready pod --timeout=30s --all; do sleep 1; done"
  timeout 210s bash -c "until ${ksh} wait --for=condition=Ready pod --timeout=30s -n kube-system --all; do sleep 1; done"
  ${ksh} get nodes
  ${ksh} get pods -A

  # Run some checks for KUBEVIRT_NUM_NODES
  # and KUBEVIRT_NUM_SECONDARY_NICS
  ${ksh} get node node01
  ${ksh} get node node02
  ${ssh} node01 -- ip l show eth1
  ${ssh} node01 -- ip l show eth2
  ${ssh} node02 -- ip l show eth1
  ${ssh} node02 -- ip l show eth2

  # Run conformance test only at CI and if the provider has them activated
  conformance_config=$DIR/${provision_dir}/conformance.json
  if [ "${CI}" == "true" -a -f $conformance_config ]; then
    hack/conformance.sh $conformance_config
  fi
)
