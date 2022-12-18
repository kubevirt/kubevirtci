#!/bin/sh
set -ex

function detect_cri() {
    if podman ps >/dev/null 2>&1; then echo "podman"; elif docker ps >/dev/null 2>&1; then echo "docker"; fi
}

export CRI_BIN=${CRI_BIN:-$(detect_cri)}

if [ "${ARCHITECTURE}" != ""  ]; then
    PLATFORM=linux/$ARCHITECTURE
fi

${CRI_BIN} run --rm --privileged multiarch/qemu-user-static --reset -p yes

if [ ! -f alpine-make-vm-image ]; then
    curl  https://raw.githubusercontent.com/alpinelinux/alpine-make-vm-image/master/alpine-make-vm-image -o alpine-make-vm-image
    chmod 755 alpine-make-vm-image
fi

${CRI_BIN} run --rm --platform=$PLATFORM -v /lib/modules:/lib/modules -v /dev:/dev --privileged -v $(pwd):$(pwd):z alpine ash -c "cd $(pwd) &&
./alpine-make-vm-image \
	--image-format qcow2 \
	--image-size 200M \
    --branch v3.16 \
	--packages \"$(cat packages)\" \
	--serial-console \
	--script-chroot \
	$1 -- configure.sh"
