#!/bin/bash

set -x

okd_image_hash="sha256:2b7b5e09b9bdf2ca40b8e153a111702584e2a3e802643e3e7df1f2d97eca0ce8"
gocli_image_hash="sha256:dd2ece308936bb13ffa040c663e30488f0130c92c0d84a7fc4f209052239747c"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd \
--random-ports \
--background \
--prefix okd-4.2 \
--registry-volume okd-4.2-registry \
--installer-secret-token ${INSTALLER_PULL_SECRET} \
"kubevirtci/okd-4.2@${okd_image_hash}"
