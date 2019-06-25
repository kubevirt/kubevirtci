#!/bin/bash

set -x

okd_image_hash="sha256:7cdb7357a7d9e8055ae2b26a9d8c926fb81440c3c5cf917407ec51297c31479f"
gocli_image_hash="sha256:caf1c3f63dc5f3137795084e492a48c802a7ef2e7f7fbe1d67e5cff684d376a3"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1.2 --registry-volume okd-4.1.2-registry "kubevirtci/okd-4.1.2@${okd_image_hash}"
