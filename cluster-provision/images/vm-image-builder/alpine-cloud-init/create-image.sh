#!/bin/sh
set -ex

KERNEL_FLAVOR="virt"
ALPINE_BRANCH="v3.19"
SCRIPT_PATH=$(dirname "$(dirname "$(realpath "$0")")")
. "${SCRIPT_PATH}/common.sh"
ARCHITECTURE="${ARCHITECTURE:-"$(go_style_local_arch)"}"
ARCH="${ARCH:-"$(linux_style_local_arch)"}"

if [ "$ARCHITECTURE" = "s390x" ]; then
   KERNEL_FLAVOR="lts"
   ALPINE_BRANCH="v3.20"
fi 

if [ "${ARCHITECTURE}" != ""  ]; then
    PLATFORM=linux/$ARCHITECTURE
fi

# s390x does not support qemu-user-static
if [ "${ARCHITECTURE}" != "s390x" ]; then
    podman run --rm --privileged docker.io/multiarch/qemu-user-static --reset -p yes
fi

if [ ! -f alpine-make-vm-image ]; then
    curl  https://raw.githubusercontent.com/kubevirt/alpine-make-vm-image/master/alpine-make-vm-image -o alpine-make-vm-image
    chmod 755 alpine-make-vm-image
fi

podman run --rm --platform=$PLATFORM -v /lib/modules:/lib/modules -v /dev:/dev --privileged -v $(pwd):$(pwd):z alpine ash -c "cd $(pwd) &&
./alpine-make-vm-image \
    --image-format qcow2 \
    --image-size 200M \
    --branch $ALPINE_BRANCH \
    --kernel-flavor $KERNEL_FLAVOR \
    --arch $ARCH \
    --packages \"$(cat packages)\" \
    --serial-console \
    --script-chroot \
    $1 -- configure.sh"
