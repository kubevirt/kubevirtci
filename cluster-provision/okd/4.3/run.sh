#!/bin/bash

set -x

okd_image_hash="sha256:63abc3884002a615712dfac5f42785be864ea62006892bf8a086ccdbca8b3d38"
gocli_image_hash="sha256:5ab9913a535227766f814b3a497d0eb1eede1a86a289d3e8bfb6bbd91836f11c"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.3 --registry-volume okd-4.3-registry "kubevirtci/okd-4.3@${okd_image_hash}"
