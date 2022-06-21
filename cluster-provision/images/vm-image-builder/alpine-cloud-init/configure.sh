#!/bin/sh

_step_counter=0
step() {
	_step_counter=$(( _step_counter + 1 ))
	printf '\n\033[1;36m%d) %s\033[0m\n' $_step_counter "$@" >&2  # bold cyan
}

step 'Set up qemu-guest-agent'
cat > /etc/conf.d/qemu-guest-agent <<-EOF
GA_METHOD="virtio-serial"
GA_PATH="/dev/virtio-ports/org.qemu.guest_agent.0"
EOF

step 'Adjust rc.conf'
sed -Ei \
	-e 's/^[# ](rc_depend_strict)=.*/\1=NO/' \
	-e 's/^[# ](rc_logger)=.*/\1=YES/' \
	-e 's/^[# ](unicode)=.*/\1=YES/' \
	/etc/rc.conf

step 'Boot without wait'
sed -Ei \
	-e "s|^[# ]*(timeout)=.*|\1=0|" \
	/etc/update-extlinux.conf

update-extlinux --warn-only 2>&1 \
    | grep -Fv 'extlinux: cannot open device /dev' >&2

step 'Enable services'
rc-update add qemu-guest-agent default
rc-update add cloud-init default
