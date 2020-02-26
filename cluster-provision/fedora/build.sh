#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

fedora_version="$(cat $DIR/version | tr -d '\n')"

docker build --build-arg fedora_version=$fedora_version . -t kubevirtci/fedora:$fedora_version
