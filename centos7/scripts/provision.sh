#!/bin/bash

set -e

NODE_NUM=${NODE_NUM-1}
n="$(printf "%02d" ${NODE_NUM})"

cat >/usr/local/bin/ssh.sh <<EOL
#!/bin/bash
set -e
dockerize -wait tcp://$(hostname -i):22${n} -timeout 300s
ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no vagrant@$(hostname -i) -i vagrant.key -p 22${n} \$@
EOL
chmod u+x /usr/local/bin/ssh.sh

if [ ! -e /dev/kvm ]; then
   set +e
   mknod /dev/kvm c 10 $(grep '\<kvm\>' /proc/misc | cut -f 1 -d' ')
   set -e
fi

# Create a transient disk, so that the container runtime does not have to copy the whole file on writes
qemu-img create -f qcow2 -o backing_file=box.qcow2 provisioned.qcow2

echo "SSH will be available via port 22${n}."
echo "VNC will be available on container port 59${n}."
echo "VM IP in the guest network will be 192.168.76.1${n}"
echo "VM hostname will be node${n}"

exec qemu-kvm -drive format=qcow2,file=provisioned.qcow2 \
  -device e1000,netdev=network0 \
  -netdev user,id=network0,hostfwd=tcp::22${n}-:22,net=192.168.66.0/24,dhcpstart=192.168.66.1${n},hostname=node${n} \
  -vnc :${n} -enable-kvm -cpu host -m 3096M
