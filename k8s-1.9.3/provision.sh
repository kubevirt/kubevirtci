#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

docker run --privileged --rm -v ${DIR}/scripts/:/scripts/ -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/cli provision --scripts /scripts --base kubevirtci/centos@sha256:94268aff21bb3b02f176b6ccbef0576d466ad31a540ca7269d6f99d31464081a --tag kubevirtci/k8s-1.9.3
