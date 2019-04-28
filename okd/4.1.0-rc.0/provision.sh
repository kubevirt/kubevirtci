#!/bin/bash

set -x

PARENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/.. && pwd )"

okd_base_hash="sha256:918d3c7f7c5ec94057715897f589c11b38e74c80927ee5af857e24817baeebaf"
gocli_image_hash="sha256:847a23412eb08217f9f062f90fd075af0f20b75e51462b1b170eba2eab7e1092"

gocli="docker run \
--privileged \
--net=host \
--rm -t \
-v /var/run/docker.sock:/var/run/docker.sock \
-v ${PARENT_DIR}:${PARENT_DIR} \
docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} provision okd \
--prefix okd-4.1.0-rc.0 \
--dir-scripts ${PARENT_DIR}/scripts \
--dir-hacks ${PARENT_DIR}/hacks \
--master-memory 10240 \
--installer-pull-token-file ${INSTALLER_PULL_SECRET} \
--installer-repo-tag release-4.1 \
--installer-release-image quay.io/openshift-release-dev/ocp-release:4.1.0-rc.0 \
"kubevirtci/okd-base@${okd_base_hash}"
