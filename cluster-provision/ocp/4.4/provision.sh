#!/bin/bash

set -ex

source $(dirname ${BASH_SOURCE[0]} )/../common.sh

tag=$(curl https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp-dev-preview/latest-4.4/release.txt |grep Name| awk '{print $2}')

# For ocp-4.4 we want OVNKubernetes
${gocli} provision okd \
--prefix ocp-4.4-provision \
--dir-scripts ${KUBEVIRTCI_PATH}/cluster-provision/okd/scripts \
--dir-manifests ${KUBEVIRTCI_PATH}/cluster-provision/manifests \
--dir-hacks ${KUBEVIRTCI_PATH}/cluster-provision/okd/hacks \
--master-memory 10240 \
--workers-memory 8192 \
--workers-cpu 4 \
--networking-type OVNKubernetes \
--installer-pull-secret-file ${INSTALLER_PULL_SECRET} \
--installer-repo-tag release-4.4 \
--installer-release-image registry.svc.ci.openshift.org/ocp/release:$tag \
"kubevirtci/okd-base@${okd_base_hash}"

$KUBEVIRTCI_PATH/cluster-provision/ocp/check-cluster-up.sh 4.4
