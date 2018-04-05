#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

docker run --privileged --rm -v ${DIR}/scripts/:/scripts/ -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/cli provision --scripts /scripts --base kubevirtci/centos@sha256:31a48682e870c6eb9a60b26e49016f654238a1cb75127f2cca37b7eda29b05e5 --tag kubevirtci/k8s-1.9.3
