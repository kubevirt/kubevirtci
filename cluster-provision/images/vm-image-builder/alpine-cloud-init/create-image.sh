#!/bin/bash -xe

domain=alpine-v3.15-$ARCHITECTURE

virsh destroy $domain || true
virsh undefine --nvram $domain || true

mkisofs -o setup.iso  setup/

qemu-img create -f qcow2 $domain.qcow2 500M

if [ "${ARCHITECTURE}" == "x86_64" ]; then
    virt_extra_args="--virt-type kvm"
else
    virt_extra_args="--arch=$ARCHITECTURE"
fi

virt-install \
    --noautoconsole \
    --name=$domain \
    --vcpus=2 \
    --memory=4096 \
    --os-type=linux \
    --os-variant=alpinelinux3.15 \
    --disk path=./alpine-virt-3.15.4-${ARCHITECTURE}.iso,device=cdrom \
    --disk path=./setup.iso,device=cdrom \
    --disk path=./$domain.qcow2,device=disk \
    --graphics none \
    --network default \
    --import \
    $virt_extra_args

DOMAIN=$domain go run setup.go

virsh destroy $domain || true

# Prepare VM image
virt-sysprep -d $domain --operations machine-id,bash-history,logfiles,tmp-files,net-hostname,net-hwaddr

# Remove VM
if [ ${ARCHITECTURE} == "x86_64" ]; then
    virsh undefine $domain
else
    virsh undefine --nvram $domain
fi

# Convert image
qemu-img convert -c -O qcow2 $domain.qcow2 $1
