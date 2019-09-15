#!/bin/bash

set -x

okd_image_hash="sha256:03b08bf66bf33c3ae1a1f63f1184761535513395e7b9c4cd496e22fc1eb2206b"
gocli_image_hash="sha256:05e7a9b04291d33f80a031438d5516cebe7ad3101108ba6702f4f67959a2fa45"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-network-4.1 --registry-volume okd-network-4.1-registry "kubevirtci/okd-network-4.1@${okd_image_hash}"
