#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source ../common-scripts/images.sh

centos_version=${IMAGES[centos7-vagrant]}

docker build --build-arg centos_version=$centos_version . -t kubevirtci/centos:$centos_version
