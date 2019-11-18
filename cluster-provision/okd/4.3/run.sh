#!/bin/bash

set -x

okd_image_hash="sha256:63abc3884002a615712dfac5f42785be864ea62006892bf8a086ccdbca8b3d38"
gocli_image_hash="sha256:f032efbd59718d48381ae2544799b44adfefa17bcd029ee6bbefdf82f53c35bc"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd \
--random-ports \
--background \
--prefix okd-4.3 \
--registry-volume okd-4.3-registry \
--installer-pull-secret-file ${INSTALLER_PULL_SECRET} \
"kubevirtci/okd-4.3@${okd_image_hash}"
