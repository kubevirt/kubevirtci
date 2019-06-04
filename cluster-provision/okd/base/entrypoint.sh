#!/usr/bin/bash

# configure iptables
iptables \
    -I INPUT \
    -p tcp \
    -s 192.168.126.0/24 \
    -d 192.168.124.1 \
    --dport 16509 \
    -j ACCEPT \
    -m comment \
    --comment "Allow insecure libvirt clients"

# start libvirt
/usr/sbin/virtlogd --daemon
/usr/sbin/libvirtd --listen
