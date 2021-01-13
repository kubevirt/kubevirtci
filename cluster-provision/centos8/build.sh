#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

centos_version="$(cat $DIR/version | tr -d '\n')"

docker build --build-arg centos_version=$centos_version . -t quay.io/kubevirtci/centos8
