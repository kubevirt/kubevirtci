#!/bin/bash

set -x

okd_image_hash="sha256:8a89ea659ffcfc6402d7d6ee43418bf2194b27ea74c239699e8268e29639aaa4"
gocli_image_hash="sha256:ac77baa6561557bb66d0f669b7aa4383430f2a7e5388627102f268f5558c3a8b"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1.0 --registry-volume okd-4.1.0-registry "kubevirtci/okd-4.1.0@${okd_image_hash}"
