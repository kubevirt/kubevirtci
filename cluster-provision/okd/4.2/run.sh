#!/bin/bash

set -x

okd_image_hash="sha256:998f79c90c635dbd8d752752c1ee1a731e40b36bbc089b00f9cdf332e6cdb72e"
gocli_image_hash="sha256:f032efbd59718d48381ae2544799b44adfefa17bcd029ee6bbefdf82f53c35bc"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd \
--random-ports \
--background \
--prefix okd-4.2 \
--registry-volume okd-4.2-registry \
--installer-pull-secret-file ${INSTALLER_PULL_SECRET} \
"kubevirtci/okd-4.2@${okd_image_hash}"
