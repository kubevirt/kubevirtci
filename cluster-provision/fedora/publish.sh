#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source ../images.sh

fedora_version=${IMAGES[fedora31-vagrant]}

docker tag kubevirtci/fedora:$fedora_version docker.io/kubevirtci/fedora:$fedora_version
docker push docker.io/kubevirtci/fedora:$fedora_version
