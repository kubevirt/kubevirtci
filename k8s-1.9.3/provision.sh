#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

docker run --privileged --rm -v ${DIR}/scripts/:/scripts/ -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/cli provision --scripts /scripts --base kubevirtci/centos@sha256:eeacdb20f0f5ec4e91756b99b9aa3e19287a6062bab5c3a41083cd245a44dc43 --tag kubevirtci/k8s-1.9.3
