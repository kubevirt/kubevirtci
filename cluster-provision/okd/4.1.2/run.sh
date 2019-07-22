#!/bin/bash

set -x

okd_image_hash="sha256:88fba5c00ba973c8da712d14689f1d93c40fa6a8e8efdb4da501b572adbd3d6b"
gocli_image_hash="sha256:1e21a7b0fc959ae2a47d1d6462ceb74f9e9181d52f9bc618e61bcab80bedaa9e"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1.2 --registry-volume okd-4.1.2-registry "kubevirtci/okd-4.1.2@${okd_image_hash}"
