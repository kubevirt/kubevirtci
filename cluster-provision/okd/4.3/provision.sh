#!/bin/bash

set -x

PARENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../.. && pwd )"

okd_base_hash="sha256:259e776998da3a503a30fdf935b29102443b24ca4ea095c9478c37e994e242bb"
gocli_image_hash="sha256:5ab9913a535227766f814b3a497d0eb1eede1a86a289d3e8bfb6bbd91836f11c"

gocli="docker run \
--privileged \
--net=host \
--rm -t \
-v /var/run/docker.sock:/var/run/docker.sock \
-v ${PARENT_DIR}:${PARENT_DIR} \
docker.io/kubevirtci/gocli@${gocli_image_hash}"

provisioner_container_id=$(docker ps --filter name=okd-4.3-provision-cluster --format {{.ID}})
docker kill $provisioner_container_id
docker container rm $provisioner_container_id

${gocli} provision okd \
--prefix okd-4.3-provision \
--dir-scripts ${PARENT_DIR}/okd/scripts \
--dir-manifests ${PARENT_DIR}/manifests \
--dir-hacks ${PARENT_DIR}/okd/hacks \
--workers-memory 8192 \
--workers-cpu 4 \
--installer-pull-token-file ${INSTALLER_PULL_SECRET} \
--installer-repo-tag release-4.3 \
--installer-release-image registry.svc.ci.openshift.org/origin/release:4.3.0-0.okd-2019-10-29-180250 \
"kubevirtci/okd-base@${okd_base_hash}"
exit $?
