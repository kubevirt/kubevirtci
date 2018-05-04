#!/bin/bash

set -ex

docker build -f "$1" . -t qemu-builder
mkdir bin
id=$(docker create qemu-builder --entrypoint /)
docker cp $id:/usr/local/bin/qemu-img "$(dirname $2)"
docker cp $id:/usr/local/bin/qemu-system-x86_64 "$(dirname $2)"
docker rm $id
