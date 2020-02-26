#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

centos_version="$(cat $DIR/version | tr -d '\n')"

docker tag kubevirtci/centos:$centos_version docker.io/kubevirtci/centos:$centos_version
docker push docker.io/kubevirtci/centos:$centos_version
