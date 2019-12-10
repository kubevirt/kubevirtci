#!/bin/bash

set -x

okd_image_hash="sha256:1fca4aae59eaaa3bec9453c4ca083cbfbddf85c63d2476caddfda04b0d5907cf"
gocli_image_hash="sha256:8dc7a694e67fadfbb337d59dfc269253079e31dca62e5298361dd464a82adc4b"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd \
--random-ports \
--background \
--prefix okd-4.1 \
--registry-volume okd-4.1-registry \
--installer-pull-secret-file ${INSTALLER_PULL_SECRET} \
"kubevirtci/okd-4.1@${okd_image_hash}"
