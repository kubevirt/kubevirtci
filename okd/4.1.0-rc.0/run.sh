#!/bin/bash

set -x

okd_image_hash="sha256:71afc34d9299aa11313b43a584e7e9d7e2f962279453b53d853a9d3bcb8b3255"
gocli_image_hash="sha256:34a5886e7d6db62f7499519fa130293a3010e86e14631a8ba80d63b29c4eb40e"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1.0-rc.0 --registry-volume okd-4.1.0-rc.0-registry "kubevirtci/okd-4.1.0-rc.0@${okd_image_hash}"
