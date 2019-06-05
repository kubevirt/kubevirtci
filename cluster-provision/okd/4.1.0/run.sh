#!/bin/bash

set -x

okd_image_hash="sha256:8a89ea659ffcfc6402d7d6ee43418bf2194b27ea74c239699e8268e29639aaa4"
gocli_image_hash="sha256:b52e44d4e44e4c03811a42af9136492fd22f725523c4a3b9258ca9556447736d"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1.0 --registry-volume okd-4.1.0-registry "kubevirtci/okd-4.1.0@${okd_image_hash}"
