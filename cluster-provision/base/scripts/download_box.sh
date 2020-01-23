#!/bin/bash

set -e
set -o pipefail

curl -L $1 | tar -zxvf - box.img
qemu-img convert -O qcow2 box.img box.qcow2
rm box.img
