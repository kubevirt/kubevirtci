#!/bin/bash

set -x

okd_image_hash="sha256:2b7b5e09b9bdf2ca40b8e153a111702584e2a3e802643e3e7df1f2d97eca0ce8"
gocli_image_hash="sha256:5ab9913a535227766f814b3a497d0eb1eede1a86a289d3e8bfb6bbd91836f11c"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1 --registry-volume okd-4.1-registry "kubevirtci/okd-4.1@${okd_image_hash}"
