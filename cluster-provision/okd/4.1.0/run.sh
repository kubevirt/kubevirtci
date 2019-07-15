#!/bin/bash

set -x

okd_image_hash="sha256:2b65ed85e98470794408df955bb1716fa7b555d3e221262030f223d1d315bfce"
gocli_image_hash="sha256:1e21a7b0fc959ae2a47d1d6462ceb74f9e9181d52f9bc618e61bcab80bedaa9e"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1.0 --registry-volume okd-4.1.0-registry "kubevirtci/okd-4.1.0@${okd_image_hash}"
