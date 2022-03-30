#!/bin/sh
set -exuo pipefail

curl -L  https://raw.githubusercontent.com/alpinelinux/alpine-make-vm-image/master/alpine-make-vm-image -o alpine-make-vm-image
chmod 755 alpine-make-vm-image

./alpine-make-vm-image --image-format qcow2 --image-size 5G \
    --repositories-file repositories \
    --packages "$(cat packages)" \
    --serial-console \
    --script-chroot $1 -- configure.sh
