#!/bin/bash

set -e
set -o pipefail


ARCH=$(uname -m)

#For the s390x architecture, instead of vagrant box image, generic cloud (qcow2) image is used directly.
if [ "$ARCH" == "s390x" ]; then
    curl -L $1 -o box.qcow2
else
    curl -L $1 | tar -zxvf - box.img
    qemu-img convert -O qcow2 box.img box.qcow2
    rm box.img
fi
