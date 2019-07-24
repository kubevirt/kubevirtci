#!/bin/bash

set -x

okd_image_hash="sha256:d452e8f910bd08b4aabe2a9b8fd82dc5984a3e95f7096b3ebd6c8ba836a5361d"
gocli_image_hash="sha256:a7880757e2d2755c6a784c1b64c64b096769ed3ccfac9d8e535df481731c2144"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1 --registry-volume okd-4.1-registry "kubevirtci/okd-4.1@${okd_image_hash}"
