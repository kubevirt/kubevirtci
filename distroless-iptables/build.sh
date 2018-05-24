#!/bin/bash

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

TAR=$1

(
  cd $DIR
  cp Dockerfile Dockerfile.iptables
  docker build -f Dockerfile.iptables . -t iptables-builder
)
id=$(docker create iptables-builder --entrypoint /)
rm -rf build
docker cp $id:/build .
docker rm $id

tar -C build -cz -O . > $TAR
