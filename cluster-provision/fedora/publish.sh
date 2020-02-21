#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

fedora_version="$(cat $DIR/version | tr -d '\n')"

docker tag kubevirtci/fedora:$fedora_version docker.io/kubevirtci/fedora:$fedora_version
docker push docker.io/kubevirtci/fedora:$fedora_version
