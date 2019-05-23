#!/bin/bash

set -x

okd_image_hash="sha256:71afc34d9299aa11313b43a584e7e9d7e2f962279453b53d853a9d3bcb8b3255"
gocli_image_hash="sha256:8677844b0c66aa02182c9d2c70b2b3bbe6f0f049e8617b630278622ca2a4228d"
okd_version="4.1.0-rc4"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-$okd_version --registry-volume okd-$okd_version-registry "kubevirtci/okd-${okd_version}@${okd_image_hash}"
