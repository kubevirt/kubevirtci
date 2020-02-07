#!/bin/bash

set -ex

source $(dirname ${BASH_SOURCE[0]} )/../common.sh


${gocli} provision okd \
--prefix ocp-4.3-provision \
--dir-scripts ${KUBEVIRTCI_PATH}/cluster-provision/okd/scripts \
--dir-manifests ${KUBEVIRTCI_PATH}/cluster-provision/manifests \
--dir-hacks ${KUBEVIRTCI_PATH}/cluster-provision/okd/hacks \
--skip-cnao \
--workers-memory 8192 \
--workers-cpu 4 \
--installer-pull-secret-file ${INSTALLER_PULL_SECRET} \
--installer-repo-tag release-4.3 \
--installer-release-image registry.svc.ci.openshift.org/ocp/release:4.3 \
"kubevirtci/okd-base@${okd_base_hash}"

$KUBEVIRTCI_PATH/cluster-provision/ocp/check-cluster-up.sh 4.3
