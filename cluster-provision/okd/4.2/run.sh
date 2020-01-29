#!/bin/bash

set -x

okd_image_hash="sha256:a830064ca7bf5c5c2f15df180f816534e669a9a038fef4919116d61eb33e84c5"
gocli_image_hash="sha256:220f55f6b1bcb3975d535948d335bd0e6b6297149a3eba1a4c14cad9ac80f80d"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd \
--random-ports \
--background \
--prefix okd-4.2 \
--registry-volume okd-4.2-registry \
--installer-pull-secret-file ${INSTALLER_PULL_SECRET} \
"kubevirtci/okd-4.2@${okd_image_hash}"
