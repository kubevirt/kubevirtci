#!/bin/bash

set -x

PARENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../.. && pwd )"

okd_base_hash="sha256:259e776998da3a503a30fdf935b29102443b24ca4ea095c9478c37e994e242bb"
gocli_image_hash="sha256:f6145018927094a6b62ac89fdb26f5901cb8030d9120f620b2490c9c9c25655a"

gocli="docker run \
--privileged \
--net=host \
--rm -t \
-v /var/run/docker.sock:/var/run/docker.sock \
-v ${PARENT_DIR}:${PARENT_DIR} \
docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} provision okd \
--prefix okd-4.1-provision \
--dir-scripts ${PARENT_DIR}/okd/scripts \
--dir-manifests ${PARENT_DIR}/manifests \
--dir-hacks ${PARENT_DIR}/okd/hacks \
--master-memory 10240 \
--installer-pull-token-file ${INSTALLER_PULL_SECRET} \
--installer-repo-tag release-4.1 \
--installer-release-image registry.svc.ci.openshift.org/rhcos-devel/vrutkovs-ocp-release:43vrutkovs.81.20191023.1 \
"kubevirtci/okd-base@${okd_base_hash}"
