#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source ../images.sh

fedora_version=${IMAGES[fedora31-vagrant]}

docker build --build-arg fedora_version=$fedora_version . -t kubevirtci/fedora:$fedora_version
