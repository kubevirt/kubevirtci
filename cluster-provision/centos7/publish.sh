#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source ../images.sh

centos_version=IMAGES[centos7-vagrant]

docker tag kubevirtci/centos:$centos_version docker.io/kubevirtci/centos:$centos_version
docker push docker.io/kubevirtci/centos:$centos_version
