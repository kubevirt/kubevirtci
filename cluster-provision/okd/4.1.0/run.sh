#!/bin/bash

set -x

okd_image_hash="sha256:d190cee4bb30e231ceb9a7c9eb1ade10c036225e126cd0abf60e9706ebd696fd"
gocli_image_hash="sha256:1e21a7b0fc959ae2a47d1d6462ceb74f9e9181d52f9bc618e61bcab80bedaa9e"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1.0 --registry-volume okd-4.1.0-registry "kubevirtci/okd-4.1.0@${okd_image_hash}"
