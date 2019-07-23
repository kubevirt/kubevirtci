#!/bin/bash

set -x

okd_image_hash="sha256:88fba5c00ba973c8da712d14689f1d93c40fa6a8e8efdb4da501b572adbd3d6b"
gocli_image_hash="sha256:868fd9f2b6e5ff6473ab92cdd7d3a762ec8a8c66dac29f5db841879d40038f2a"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1.2 --registry-volume okd-4.1.2-registry "kubevirtci/okd-4.1.2@${okd_image_hash}"
