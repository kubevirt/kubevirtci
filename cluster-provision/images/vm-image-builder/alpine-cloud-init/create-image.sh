#!/bin/sh
set -ex

# Currently, the build tool, alpine-make-vm-image, only support amd64
# So we disable build for non-x86 architectures
# https://github.com/alpinelinux/alpine-make-vm-image/issues/10
if [[ ${ARCHITECTURE} != "" && ${ARCHITECTURE} != "amd64" || $(uname -m) != "x86_64" ]]; then
   echo "only support native build for amd64 platform"
   exit 1
fi

if [ "${ARCHITECTURE}" != ""  ]; then
    PLATFORM=linux/$ARCHITECTURE
fi

podman run --rm --privileged docker.io/multiarch/qemu-user-static --reset -p yes

if [ ! -f alpine-make-vm-image ]; then
    curl  https://raw.githubusercontent.com/alpinelinux/alpine-make-vm-image/master/alpine-make-vm-image -o alpine-make-vm-image
    chmod 755 alpine-make-vm-image
fi

podman run --rm --platform=$PLATFORM -v /lib/modules:/lib/modules -v /dev:/dev --privileged -v $(pwd):$(pwd):z alpine ash -c "cd $(pwd) &&
./alpine-make-vm-image \
    --image-format qcow2 \
    --image-size 200M \
    --branch v3.19 \
    --packages \"$(cat packages)\" \
    --serial-console \
    --script-chroot \
    $1 -- configure.sh"
