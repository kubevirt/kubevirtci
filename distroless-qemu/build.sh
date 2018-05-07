#!/bin/bash

set -ex

DOCKERFILE=$1
TAR=$2

rm -rf build
docker build -f "$DOCKERFILE" . -t qemu-builder
id=$(docker create qemu-builder --entrypoint /)
docker cp $id:/build .
docker rm $id
tar -C build -cz -O . > $TAR
