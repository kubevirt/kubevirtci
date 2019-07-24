#!/bin/bash

set -x

okd_image_hash="sha256:d190cee4bb30e231ceb9a7c9eb1ade10c036225e126cd0abf60e9706ebd696fd"
gocli_image_hash="sha256:868fd9f2b6e5ff6473ab92cdd7d3a762ec8a8c66dac29f5db841879d40038f2a"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1.0 --registry-volume okd-4.1.0-registry "kubevirtci/okd-4.1.0@${okd_image_hash}"
