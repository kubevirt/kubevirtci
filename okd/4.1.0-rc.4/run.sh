#!/bin/bash

set -x

okd_image_hash="sha256:0fda220dca5569230d507e68a80d8ff9c6d34a778f4178f3a1a316137408c609"
gocli_image_hash="sha256:8cc901ae21608d6d1689d7e70aa5e69889f3f945d52f1b140fd8286428a0bc3a"
okd_version="4.1.0-rc.4"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-$okd_version --registry-volume okd-$okd_version-registry "kubevirtci/okd-${okd_version}@${okd_image_hash}"
