#!/bin/bash

set -x

okd_image_hash="sha256:8a89ea659ffcfc6402d7d6ee43418bf2194b27ea74c239699e8268e29639aaa4"
gocli_image_hash="sha256:caf1c3f63dc5f3137795084e492a48c802a7ef2e7f7fbe1d67e5cff684d376a3"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1.0 --registry-volume okd-4.1.0-registry "kubevirtci/okd-4.1.0@${okd_image_hash}"
