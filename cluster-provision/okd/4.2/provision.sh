#!/bin/bash

set -x

PARENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../.. && pwd )"

okd_base_hash="sha256:259e776998da3a503a30fdf935b29102443b24ca4ea095c9478c37e994e242bb"
gocli_image_hash="sha256:a7880757e2d2755c6a784c1b64c64b096769ed3ccfac9d8e535df481731c2144"

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
--workers-memory 8192 \
--workers-cpu 4 \
--installer-pull-token-file ${INSTALLER_PULL_SECRET} \
--installer-repo-tag release-4.2 \
--installer-release-image docker.io/kubevirtci/ocp-release@sha256:79347c0c3412f0bcc99af573d15ed60ed670ec6c287f98267c3e22ffcf577370 \
"kubevirtci/okd-base@${okd_base_hash}"
