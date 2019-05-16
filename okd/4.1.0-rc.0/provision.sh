#!/bin/bash

set -x

PARENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../.. && pwd )"

okd_base_hash="sha256:918d3c7f7c5ec94057715897f589c11b38e74c80927ee5af857e24817baeebaf"
gocli_image_hash="sha256:34a5886e7d6db62f7499519fa130293a3010e86e14631a8ba80d63b29c4eb40e"

gocli="docker run \
--privileged \
--net=host \
--rm -t \
-v /var/run/docker.sock:/var/run/docker.sock \
-v ${PARENT_DIR}:${PARENT_DIR} \
docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} provision okd \
--prefix okd-4.1.0-rc.0 \
--dir-scripts ${PARENT_DIR}/okd/scripts \
--dir-manifests ${PARENT_DIR}/manifests \
--dir-hacks ${PARENT_DIR}/okd/hacks \
--master-memory 10240 \
--installer-pull-token-file ${INSTALLER_PULL_SECRET} \
--installer-repo-tag release-4.1 \
--installer-release-image quay.io/openshift-release-dev/ocp-release:4.1.0-rc.0 \
"kubevirtci/okd-base@${okd_base_hash}"
