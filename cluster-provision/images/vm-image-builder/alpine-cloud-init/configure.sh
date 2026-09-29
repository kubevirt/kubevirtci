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

step 'Install udhcpc6 script'
cat udhcpc6.script > /etc/udhcpc6.script
chmod 755 /etc/udhcpc6.script

step 'Configure IPv6 DHCP on eth0'
cat >> /etc/network/interfaces <<'EOF'

# IPv6 (explicitly calls udhcpc6 on link UP)
iface eth0 inet6 manual
    up /usr/bin/udhcpc6 -b -t 3 -p /var/run/udhcpc6.eth0.pid -i eth0 -s /etc/udhcpc6.script
    down kill $(cat /var/run/udhcpc6.eth0.pid 2>/dev/null) 2>/dev/null || true
EOF

step 'Enable services'
rc-update add syslog boot
rc-update add qemu-guest-agent default
rc-update add cloud-init default
rc-update add cloud-init-local default
rc-update add cloud-config default
rc-update add cloud-final default
rc-update add networking default
rc-update add sshd default
rc-update add acpid default
rc-update add udev sysinit
rc-update add udev-trigger sysinit
rc-update add udev-settle sysinit
rc-update add udev-postmount default
