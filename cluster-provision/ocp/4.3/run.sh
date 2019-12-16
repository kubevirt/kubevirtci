#!/bin/bash

set -x

ocp_image_hash="sha256:16a70403141142aae387a50feb2fd039a745c6916aa3f61e1a5d5a74efb6be39"
gocli_image_hash="sha256:a7880757e2d2755c6a784c1b64c64b096769ed3ccfac9d8e535df481731c2144"

gocli="docker run --privileged --net=host --rm -t -v /var/run/docker.sock:/var/run/docker.sock docker.io/kubevirtci/gocli@${gocli_image_hash}"

${gocli} run ocp --random-ports --background --prefix ocp-4.3 --registry-volume ocp-4.3-registry "kubevirtci/ocp-4.3@${ocp_image_hash}"
