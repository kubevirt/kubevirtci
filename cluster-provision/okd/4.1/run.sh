#!/bin/bash

set -x

okd_image_hash="sha256:67fe42feea8256f07069d776d4c4cecff6294ff8a5af67d719eca6c08548b45d"
gocli_image_hash="sha256:8dc7a694e67fadfbb337d59dfc269253079e31dca62e5298361dd464a82adc4b"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd \
--random-ports \
--background \
--prefix okd-4.1 \
--registry-volume okd-4.1-registry \
--installer-pull-secret-file ${INSTALLER_PULL_SECRET} \
"kubevirtci/okd-4.1@${okd_image_hash}"
