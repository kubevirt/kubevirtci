#!/bin/bash

set -x

okd_image_hash="sha256:03b08bf66bf33c3ae1a1f63f1184761535513395e7b9c4cd496e22fc1eb2206b"
gocli_image_hash="sha256:a7880757e2d2755c6a784c1b64c64b096769ed3ccfac9d8e535df481731c2144"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.2 --registry-volume okd-4.2-registry "kubevirtci/okd-4.2@${okd_image_hash}"
