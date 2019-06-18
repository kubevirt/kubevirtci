#!/bin/bash

set -x

okd_image_hash="sha256:7cdb7357a7d9e8055ae2b26a9d8c926fb81440c3c5cf917407ec51297c31479f"
gocli_image_hash="sha256:b52e44d4e44e4c03811a42af9136492fd22f725523c4a3b9258ca9556447736d"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1.2 --registry-volume okd-4.1.2-registry "kubevirtci/okd-4.1.2@${okd_image_hash}"
