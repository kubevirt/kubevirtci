#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

docker tag kubevirtci/centos8 docker.io/kubevirtci/centos8
docker push docker.io/kubevirtci/centos8
