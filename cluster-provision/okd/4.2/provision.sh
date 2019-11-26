#!/bin/bash

set -x

PARENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../.. && pwd )"

okd_base_hash="sha256:73ede51ce464546a82b81956b7f58cf98662a4c5fded9c659b57746bc131e047"
gocli_image_hash="sha256:f032efbd59718d48381ae2544799b44adfefa17bcd029ee6bbefdf82f53c35bc"

gocli="docker run \
--privileged \
--net=host \
--rm -t \
-v /var/run/docker.sock:/var/run/docker.sock \
-v ${PARENT_DIR}:${PARENT_DIR} \
docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} provision okd \
--prefix okd-4.2-provision \
--dir-scripts ${PARENT_DIR}/okd/scripts \
--dir-manifests ${PARENT_DIR}/manifests \
--dir-hacks ${PARENT_DIR}/okd/hacks \
--master-memory 10240 \
--workers-cpu 4 \
--workers-memory 6144 \
--installer-pull-secret-file ${INSTALLER_PULL_SECRET} \
--installer-repo-tag release-4.2 \
--installer-release-image docker.io/kubevirtci/ocp-release:4.2.5 \
"kubevirtci/okd-base@${okd_base_hash}"
