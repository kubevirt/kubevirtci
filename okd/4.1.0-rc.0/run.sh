#!/bin/bash

set -x

okd_image_hash="sha256:7b5badcf46fceae521706161622ee447608d8ff84f831729ea98e061da49f4d7"
gocli_image_hash="sha256:847a23412eb08217f9f062f90fd075af0f20b75e51462b1b170eba2eab7e1092"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run okd --random-ports --background --prefix okd-4.1.0-rc.0 --registry-volume okd-4.1.0-rc.0-registry "kubevirtci/okd-4.1.0-rc.0@${okd_image_hash}"
