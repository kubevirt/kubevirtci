#!/bin/bash

set -x

okd_image_hash="sha256:76b487d894ab89a91ba4985591a7ff05e91be9665face1492c23405aad2d0201"
gocli_image_hash="sha256:9ca5dfbb4c4e70aefcc63942073808037a39eec3394bb6f9416d4c1ca3340997"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1 --registry-volume okd-4.1-registry "kubevirtci/okd-4.1@${okd_image_hash}"
