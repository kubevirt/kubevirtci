#!/bin/bash

set -x

PARENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../.. && pwd )"

okd_base_hash="sha256:918d3c7f7c5ec94057715897f589c11b38e74c80927ee5af857e24817baeebaf"
gocli_image_hash="sha256:8677844b0c66aa02182c9d2c70b2b3bbe6f0f049e8617b630278622ca2a4228d"
okd_version="4.1.0-rc.4"

gocli="docker run \
--privileged \
--net=host \
--rm -t \
-v /var/run/docker.sock:/var/run/docker.sock \
-v ${PARENT_DIR}:${PARENT_DIR} \
docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} provision okd \
--prefix okd-${okd_version} \
--dir-scripts ${PARENT_DIR}/okd/scripts \
--dir-manifests ${PARENT_DIR}/manifests \
--dir-hacks ${PARENT_DIR}/okd/hacks \
--master-memory 10240 \
--installer-pull-token-file ${INSTALLER_PULL_SECRET} \
--installer-repo-commit 187ce5e2c6fb4878b170d86cef7ecf1f50fea70f \
--installer-release-image quay.io/openshift-release-dev/ocp-release:${okd_version} \
"kubevirtci/okd-base@${okd_base_hash}"
