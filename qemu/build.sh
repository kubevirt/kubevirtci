#!/bin/bash

set -ex

docker build . -t qemu-builder
id=$(docker create qemu-builder --entrypoint /)
docker cp $id:/usr/local/bin/qemu-io .
docker cp $id:/usr/local/bin/qemu-system-x86_64 .
docker rm $id
