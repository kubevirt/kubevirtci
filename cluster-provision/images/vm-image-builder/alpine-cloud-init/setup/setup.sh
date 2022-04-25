#!/bin/sh -xe

_step_counter=0
step() {
	_step_counter=$(( _step_counter + 1 ))
	printf '\n\033[1;36m%d) %s\033[0m\n' $_step_counter "$@" >&2  # bold cyan
}

step 'Set up networking with DHCP and restart interfaces'  
setup-interfaces -a -r

step 'Configure apk with the first mirror and community repos'
setup-apkrepos -1 -c

step 'Setup ssh daemon'
setup-sshd -c openssh

step 'Installing ga and cloud-init packages'
apk add qemu-guest-agent cloud-init e2fsprogs-extra util-linux

step 'Set up qemu-guest-agent'
cat > /etc/conf.d/qemu-guest-agent <<-EOF
GA_METHOD="virtio-serial"
GA_PATH="/dev/vport1p1"
EOF

step 'Adjust rc.conf'
sed -Ei \
	-e 's/^[# ](rc_depend_strict)=.*/\1=NO/' \
	-e 's/^[# ](rc_logger)=.*/\1=YES/' \
	-e 's/^[# ](unicode)=.*/\1=YES/' \
	/etc/rc.conf

step 'Enable services'
rc-update add cloud-final default
rc-update add cloud-config default
rc-update add cloud-init default
setup-udev -n
rc-update add qemu-guest-agent default

step 'Set up system disk'
ERASE_DISKS=/dev/vda setup-disk -m sys /dev/vda
