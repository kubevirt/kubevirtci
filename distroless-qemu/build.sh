#!/bin/bash

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

TAR=$1

(
  cd $DIR
  cp Dockerfile Dockerfile.qemu
  docker build -f Dockerfile.qemu . -t qemu-builder
)
id=$(docker create qemu-builder --entrypoint /)
rm -rf build
docker cp $id:/build .
docker rm $id

tar -C build -cz -O . > $TAR
