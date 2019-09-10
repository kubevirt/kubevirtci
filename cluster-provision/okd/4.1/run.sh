#!/bin/bash

set -x

okd_image_hash="sha256:fb093cb0e9f06584d0d0bb993e7242a463e25d80c1d789659dcf4c8b3d91a5af"
gocli_image_hash="sha256:a7880757e2d2755c6a784c1b64c64b096769ed3ccfac9d8e535df481731c2144"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1 --registry-volume okd-4.1-registry "kubevirtci/okd-4.1@${okd_image_hash}"
