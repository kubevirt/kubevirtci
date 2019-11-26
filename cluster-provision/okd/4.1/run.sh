#!/bin/bash

set -x

okd_image_hash="sha256:67fe42feea8256f07069d776d4c4cecff6294ff8a5af67d719eca6c08548b45d"
gocli_image_hash="sha256:f032efbd59718d48381ae2544799b44adfefa17bcd029ee6bbefdf82f53c35bc"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd \
--random-ports \
--background \
--prefix okd-4.1 \
--registry-volume okd-4.1-registry \
--installer-pull-secret-file ${INSTALLER_PULL_SECRET} \
"kubevirtci/okd-4.1@${okd_image_hash}"
